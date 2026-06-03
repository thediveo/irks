# irks

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/irks)
[![License](https://img.shields.io/github/license/thediveo/irks)](https://img.shields.io/github/license/thediveo/irks)
![build and test](https://github.com/thediveo/irks/actions/workflows/buildandtest.yaml/badge.svg?branch=master)
![Coverage](https://img.shields.io/badge/Coverage-94.5%25-brightgreen)

`irks` is a Go module for retrieving IRQ counters, structure, and CPU affinity.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## DevContainer

> [!CAUTION]
>
> Do **not** use VSCode's "~~Dev Containers: Clone Repository in Container
> Volume~~" command, as it is utterly broken by design, ignoring
> `.devcontainer/devcontainer.json`.

1. `git clone https://github.com/thediveo/irks`
2. in VSCode: Ctrl+Shift+P, "Dev Containers: Open Workspace in Container..."
3. select `irks.code-workspace` and off you go...

## Go Version Support

`notwork` supports versions of Go that are noted by the Go release policy, that
is, major versions _N_ and _N_-1 (where _N_ is the current major version).

## Copyright and License

`irks` is Copyright 2024, 2026 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
