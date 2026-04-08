# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gomock is a lightweight HTTP mock server written in Go. It serves mocked JSON responses based on a JSON configuration file, supporting static responses, proxying, dynamic read/write storage, CORS, configurable error simulation, and hot-reload of config files via fsnotify.

## Development process

- `make checkfmt`, `make lint` and `make test` must pass.
- always make sure to add tests for any new code you write.
- tests should be placed next to the code they test.

## Common Commands

- **Build:** `make build` (outputs to `_dist/`, defaults to darwin/arm64)
- **Run:** `make run` (runs with `_dist/mock.json`, watch mode, verbose)
- **Run directly:** `go run cmd/app/main.go -mock <path-to-mock.json> [-watch] [-verbose]`
- **Test:** `make test` (runs `gofmt`, `staticcheck`, `go test -v ./...` with coverage)
- **Run a single test:** `go test -v -run TestFunctionName ./app/...`
- **Install locally:** `make install` (builds and copies binary to `~/bin/gomock`)

## Architecture

The entrypoint is `cmd/app/main.go`. It loads two configs:
- **App config** (`_resources/config.yml` via `-config` flag): YAML file for server settings (timeouts, address, log level). Parsed in `config/config.go`.
- **Mock config** (`mock.json` via `-mock` flag): JSON file defining endpoint behaviors. Parsed in `config/mock.go` into `config.Mock` / `config.Endpoint` structs.

The main loop supports hot-reload: when `-watch` is set, file changes trigger server restart via context cancellation.

`app/` contains the HTTP layer:
- `app.go` — Registers routes on a `chi.Mux` router, defines `appHandler`/`appError` pattern for consistent error handling.
- `handlers.go` — `setupAPI()` iterates mock endpoints and registers chi routes. `apiHandler()` is the core handler implementing: delay simulation, sampled error injection, proxying, static JSON responses, `jsonPath` file serving, and dynamic store read/write.
- `store.go` — In-memory key-value store (`sync.RWMutex`-based) for dynamic mock data.
- `proxy.go` — Reverse proxy wrapper.
- `cors.go` — CORS middleware.

## Key Dependencies

- `go-chi/chi/v5` — HTTP router (replaced gorilla/mux; README references to gorilla mux path syntax are outdated)
- `go.uber.org/zap` — Structured logging
- `gopkg.in/fsnotify.v1` — File watching for hot-reload
- `stretchr/testify` — Test assertions
