# Go Mock

[![Build Status](https://travis-ci.com/smeshkov/gomock.svg?branch=master)](https://travis-ci.com/smeshkov/gomock)

Light weight mock server written in [Go](https://golang.org/).

# Usage

Put `api.json` near to `gomock` binary and start mock server by running `./gomock`.

`api.json` defines mocks - handlers which will provided predefined JSON responses.

`api.json` example:

```json
{
  "port": 3000, // optional (defualts to 8080)
  "endpoints": [
    {
      "method": "GET", // optional (defualts to GET)
      "status": 200, // optional (defualts to GET)
      "path": "/users",
      "jsonPath": "./users.json" // one way of defining response payload: will read from ./users.json to response
    },
    {
      // "method" and "status" are omited, therefore set to defaults GET and 200
      "path": "/user/1",
      "json": { // another way of defining response payload: will output given JSON
        "id": 1,
        "name": "name",
        "address": "address"
      }
    },
    {
      "method": "POST",
      "path": "/error",
      "status": 400,
      "json": {
        "message": "failed"
      }
    }
  ]
}
```

## Changelog

See [CHANGELOG.md](https://raw.githubusercontent.com/smeshkov/lsh/master/CHANGELOG.md)

## License

Released under the [Apache License 2.0](https://raw.githubusercontent.com/smeshkov/gomock/master/LICENSE).
