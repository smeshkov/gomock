# Changelog

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