# Development instructions

- Use Go 1.26.10 for this repository. Before any local Go command, source `~/.gvm/scripts/gvm` and run `gvm use go1.26.10`; do not rely on `/usr/local/go/bin/go`.
- Keep `go.mod`, the Docker builder and the README development baseline aligned. CI reads its Go version from `go.mod`.
- Validate with `go mod verify`, `go build ./...`, `go test -race -count=1 ./...` and `go vet ./...`.
- Notification tests must use local HTTP servers or mock senders. Never send real Slack messages, email or production alerts during validation.
