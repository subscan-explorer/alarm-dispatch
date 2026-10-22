# alarm-dispatch
alertmanager alarm dispatch

## Development

Use Go **1.26.10** for local builds, CI and the Docker builder. This includes the
`crypto/x509` fix for [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809).
`golang.org/x/net` is pinned to v0.60.0, which requires Go 1.26.

```sh
source ~/.gvm/scripts/gvm
gvm use go1.26.10
export GOTOOLCHAIN=local
go mod download
go mod verify
go build ./...
go test -race -count=1 ./...
go vet ./...
make build
```

The notification regression tests use a local webhook and a mock email sender;
they do not require notification credentials. Use these tests for validation,
without starting the service against a production alert configuration.

CI reads the Go version from `go.mod`. Keep the Docker builder and these local
development instructions aligned whenever the Go security baseline changes.
