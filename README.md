
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
| macOS    | amd64 (Intel) | O           | O            |
|          | arm64 (Apple) | O           | O            |
| Windows  | amd64         | imminent    | X            |     

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
        "build.buildFlags": [
            "-tags=edge_edition"
        ]
    }
```
