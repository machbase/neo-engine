
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
- [neo-grpc](https://github.com/machbase/neo-grpc) gRPC interface
- [neo-spi](https://github.com/machbase/neo-spi) Defines the general interface
- [neo-web](https://github.com/machbase/neo-web) machbase-neo web ui

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
#### VSCode settings.json for Windows

- `./vscode/settings.json`

```json
{
    "files.exclude": {
        "**/.git": true,
        "**/.svn": true,
        "**/.hg": true,
        "**/CVS": true,
        "**/.DS_Store": true,
        "**/Thumbs.db": true,
        "**/vendor": true
    },
    "gopls": {
        "ui.semanticTokens": true,
        "ui.completion.usePlaceholders": true,
        "buildFlags": ["-tags", "fog_edition"]
    },
    "go.toolsEnvVars": {
        "GOOS": "windows",
        "GOARCH": "amd64",
        "CGO_ENABLED":"1",
        "CC": "C:\\TDM-GCC-64\\bin\\gcc.exe",
        "CXX": "C:\\TDM-GCC-64\\bin\\g++.exe"
    },
    "go.testFlags": ["-timeout", "60s", "-v", "-count=1", "-race", "-cover", "-tags=fog_edition"]
}
```