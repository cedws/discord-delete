BIN_DIR=bin
BIN_NAME=discord-delete
GOFLAGS=-trimpath

all: clean test build

clean:
	rm -rf $(BIN_DIR)

test:
	go test -failfast -race ./...

build: build-linux build-darwin build-windows

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-arm64

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-darwin-amd64

build-windows:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe

.PHONY: clean test build build-linux build-darwin build-windows
