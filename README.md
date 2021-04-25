# Go Mock

Light weight mock server written in [Go](https://golang.org/).

## Installation

For MacOS:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/smeshkov/gomock/master/_bin/install.sh)"
```

For Linux:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/smeshkov/gomock/master/_bin/install.sh linux)"
```

## Usage

1. [Download](https://github.com/smeshkov/gomock/releases/latest) latest release.
2. Put `mock.json` near to `gomock` binary and start mock server by running `./gomock_darwin_v<version>` (or `./gomock_linux_v<version>`).

Use `./gomock_darwin_v<version> --help` (or `./gomock_linux_v<version> --help`) for help.

`mock.json` defines mocks - handlers which will provided predefined JSON responses.

`mock.json` example:

```json
{
  "port": 8080,
  "endpoints": [
    {
      "methods": [ "GET" ],
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
      },
      "allowCors": [ "example.com" ]
    },
    {
      "methods": ["POST"],
      "path": "/error",
      "delay": 100,
      "status": 400,
      "json": {
        "message": "failed"
      }
    },
    {
      "methods": [ "POST" ],
      "path": "/api/*",
      "delay": 200,
      "url": "http://localhost:8090"
    }
  ]
}
```

Mock JSON configuration properties:

- `port` - optional (defaults to 8080);
- `endpoints` - an array of endpoints to configure;

Endpoint object in `endpoints` list:

- `methods` - list of allowed methods, optional defaults to "GET";
- `path` - URL path to the mocked endpoint, if not set, then defaults to catch all;
- `delay` - delay in milliseconds on the server side;
- `status` - HTTP response status code, optional defaults to 200;
- `json` - one way of defining response payload, will output given JSON;
- `jsonPath` - another way of defining response payload, will read file from the given path (can be relative to the root mock JSON file) and write its contents to response;
- `proxy` - proxies requests to the given address;
- `errors` - helps to setup sampled errors, with the randomised error codes.
- `allowCors` - list of allowed domains for CORS.

`mock.json` is the default name for a mock configuration file, it can be renamed and set via `-mock` option, e.g. `./gomock -mock api.json`

## Changelog

See [CHANGELOG.md](https://raw.githubusercontent.com/smeshkov/gomock/master/CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](https://raw.githubusercontent.com/smeshkov/gomock/master/CONTRIBUTING.md)

## License

Released under the [Apache License 2.0](https://raw.githubusercontent.com/smeshkov/gomock/master/LICENSE).
