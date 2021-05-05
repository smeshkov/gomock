# Go Mock

Light weight mock server written in [Go](https://golang.org/). 

Supported features:

* good old mocked JSON responses for HTTP methods and URI paths;
* custom error behaviours;
* proxying;
* dynamic results, as in store data from an incoming JSON and then retrieve it.

## Installation

For MacOS:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/smeshkov/gomock/master/_bin/install.sh)"
```

For Linux:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/smeshkov/gomock/master/_bin/install.sh)" linux
```

For Windows:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/smeshkov/gomock/master/_bin/install.sh)" windows
```

## Usage

Create JSON file with your mocks and start mock server by running `gomock -mock /path/to/your/mock.json`.

`mock.json` defines mocks - handlers which will serve provided mocks.

Use `gomock --help` for more information.

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
- `dynamic` - allows to configure dynamic read/write behaviour, i.e. values can be stored and retrieved from the inetrnal store.

`mock.json` is the default name for a mock configuration file, it can be renamed and set via `-mock` option, e.g. `./gomock -mock api.json`

## Dynamic mocking

You can store and retrieve values in your mocks by using `dynamic` property.

For writes use `dynamic.write.json`:

```json
{
  "port": 8080,
  "endpoints": [
    {
      "methods": [ "POST" ],
      "path": "/note",
      "dynamic": {
        "write": {
          "json": {
            "name": "note",
            "key": "/id", // path to an entity's key inside the incoming request JSON from the client ("id" field in this case)
            "value": "." // path to an entity's value inside the incoming request JSON from the client (root in this case)
          }
        }
      }
    }
  ]
}
```

For reads use `dynamic.read.json`:

```json
{
  "port": 8080,
  "endpoints": [
    {
      "methods": [ "GET" ],
      "path": "/note/{noteID:[a-zA-Z0-9-]+}", // uses Gorilla Mux paths
      "dynamic": {
        "read": {
          "json": {
            "name": "note",
            "keyParam": "noteID", // path to an entity's key inside the incoming request path from the client ("noteID" param in this case)
          }
        }
      }
    }
  ]
}
```

## Changelog

See [CHANGELOG.md](https://raw.githubusercontent.com/smeshkov/gomock/master/CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](https://raw.githubusercontent.com/smeshkov/gomock/master/CONTRIBUTING.md)

## License

Released under the [Apache License 2.0](https://raw.githubusercontent.com/smeshkov/gomock/master/LICENSE).
