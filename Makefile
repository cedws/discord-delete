BIN_DIR  = bin
BIN_NAME = discord-delete
GOFLAGS  = -trimpath

all: clean test build

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

.PHONY: test
test:
	go test -failfast -race ./...

.PHONY: build
build: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-arm64

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-darwin-amd64

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe
