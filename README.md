# irks

![Coverage](https://img.shields.io/badge/Coverage-98.2%25-brightgreen)

`irks` is a Go module for retrieving IRQ counters, structure, and CPU affinity.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Go Version Support

`notwork` supports versions of Go that are noted by the Go release policy, that
is, major versions _N_ and _N_-1 (where _N_ is the current major version).

## Make Targets

- `make`: lists all targets.
- `make coverage`: runs all tests with coverage and then **updates the coverage
  badge in `README.md`**.
- `make pkgsite`: installs [`x/pkgsite`](https://golang.org/x/pkgsite/cmd/pkgsite), as
  well as the [`browser-sync`](https://www.npmjs.com/package/browser-sync) and
  [`nodemon`](https://www.npmjs.com/package/nodemon) npm packages first, if not
  already done so. Then runs the `pkgsite` and hot reloads it whenever the
  documentation changes.
- `make report`: installs
  [`@gojp/goreportcard`](https://github.com/gojp/goreportcard) if not yet done
  so and then runs it on the code base.
- `make test`: runs **all** tests, always.
- `make vuln`: installs
  [`x/vuln/cmd/govulncheck`](https://golang.org/x/vuln/cmd/govulncheck) and then
  runs it.

## Copyright and License

`irks` is Copyright 2024 Harald Albrecht, and licensed under the Apache License,
Version 2.0.
