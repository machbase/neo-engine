
[![CI linux-amd64](https://github.com/machbase/neo-engine/actions/workflows/ci-linux-amd64.yml/badge.svg)](https://github.com/machbase/neo-engine/actions/workflows/ci-linux-amd64.yml)
[![CI darwin-amd64](https://github.com/machbase/neo-engine/actions/workflows/ci-darwin-amd64.yml/badge.svg)](https://github.com/machbase/neo-engine/actions/workflows/ci-darwin-amd64.yml)

# neo-engine

Go binding for Machbase time-series database core.

## Install

```sh
go get -u github.com/machbase/neo-engine
```

## Supporting platforms

| OS       | ARCH          | fog_edition | edge_edition |
|:---------|:--------------|-------------|--------------|
| Linux    | amd64         | O           | O            |
|          | arm64         | O           | O            |
|          | x86 (32bit)   | X           | imminent     |
| macOS    | amd64 (Intel) | imminent    | imminent     |
|          | arm64 (Apple) | O           | O            |
| Windows  | amd64         | imminent    | X            |     

## Related projects

- [neo-server](https://github.com/machbase/neo-server) machbase-neo server
- [neo-shell](https://github.com/machbase/neo-shell) machbase-neo shell
- [neo-grpc](https://github.com/machbase/neo-grpc) gRPC interface
- [neo-spi](https://github.com/machbase/neo-spi) Defines the general interface

## Development environment

### VSCode Build flags

Set one of editions as Go build tags

- edge_edition
- fog_edition

#### Command line

```sh
go build -tags edge_edition
```

#### VSCode settings.json

```json
    "gopls": {
        "buildFlags": ["-tags", "fog_edition"]
    }
```
