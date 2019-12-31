# Go Mock

[![Build Status](https://travis-ci.com/smeshkov/gomock.svg?branch=master)](https://travis-ci.com/smeshkov/gomock)

Light weight mock server written in [Go](https://golang.org/).

# Usage

1. [Download](https://github.com/smeshkov/gomock/releases/latest) latest release.
2. Put `mock.json` near to `gomock` binary and start mock server by running `./gomock_darwin_v<version>` (or `./gomock_linux_v<version>`).

Use `./gomock_darwin_v<version> --help` (or `./gomock_linux_v<version> --help`) for help.

`mock.json` defines mocks - handlers which will provided predefined JSON responses.

`mock.json` example:

```json
{
  "port": 3000,
  "endpoints": [
    {
      "method": "GET",
      "status": 200,
      "path": "/users",
      "jsonPath": "./users.json"
    },
    {
      "path": "/user/1",
      "json": {
        "id": 1,
        "name": "name",
        "address": "address"
      }
    },
    {
      "method": "POST",
      "path": "/error",
      "delay": 100,
      "status": 400,
      "json": {
        "message": "failed"
      }
    },
    {
      "method": "POST",
      "path": "*",
      "delay": 200,
      "url": "http://localhost:3003"
    }
  ]
}
```

Mock configuration fields:

- `port` - optional (defaults to 8080);
- `endpoints` - an array of endpoints to configure;

Endpoint object in `endpoints` list:

- `method` - optional (defaults to "GET");
- `path` - required, URL path to the mocked endpoint;
- `delay` - delay in milliseconds on the server side;
- `status` - optional (defaults to 200);
- `json` - one way of defining response payload, will output given JSON;
- `jsonPath` - another way of defining response payload, will read file from the given path and write its contents to response;
- `proxy` - proxies requests to the given address;
- `errors` - helps to setup sampled errors, with the randomised error codes.

mock.json is the defaukt name for mock configuration file, it can be renamed and set via `-mock` option, e.g. `./gomock -mock api.json`

## Changelog

See [CHANGELOG.md](https://raw.githubusercontent.com/smeshkov/lsh/master/CHANGELOG.md)

## License

Released under the [Apache License 2.0](https://raw.githubusercontent.com/smeshkov/gomock/master/LICENSE).
