.PHONY: check run build

BUILD_PATH=cmd

export CGO_ENABLED=0

getdeps:
	@mkdir -p $(GOPATH)/bin
	@which golangci-lint 1>/dev/null || (echo "Installing golangci-lint" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

lint: getdeps
	@echo "Running $@ check"
	${GOPATH}/bin/golangci-lint cache clean
	${GOPATH}/bin/golangci-lint run --timeout=5m --config ./.golangci.yml

check: lint

run:
	@go run ./$(BUILD_PATH)

build:
	@go build -o ./bin/alarm-bot -v ./$(BUILD_PATH)