# Changelog

## v0.11.0

- added watch config implemented by @mrchypark;
- fixed logger for hotreload (thanks to @mrchypark);
- updated `ioutile` to `os` (thanks to @mrchypark).

## v0.10.1

- fixed bug, with the unnecessary `url.QueryEscape` for the query params when proxying requests.

## v0.10.0

- handle more complex patterns in the paths like `/foo*`;
- bumped Go version to 1.23.

## v0.9.4

- added darwin arm64 binary;
- release for Go 1.20;
- removed use of deprecated "io/ioutil".

## v0.9.0

- migrated from Gorilla to Chi for HTTP handling;
- introduced `static` property for an endpoint, to serve static files;
- bumped Go to 1.20.

## v0.8.0

- added support for subpath wildcards.

## v0.7.0

- added `-version` option.

## v0.6.0

- added `dynamic` endpoint property to dynamically store and retrieve values from your mocks;
- improved `proxy` mode, so it actually works now;
- added `-verbose` option.

## v0.5.0

- added `allowCors` endpoint property to configure CORS access;
- `jspnPath` now correctly resolves files realtive to the root mock JSON file.

## v0.4.0

- Bumped Go version to 1.16.

## v0.3.0

 - Added `errors` in endpoints to be able to setup sampled erroring;
 - Added `proxy` in endpoints for proxying calls to another address;
 - Moved onto `go.uber.org/zap`.

## v0.2.0

 - First release.