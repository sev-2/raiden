CLI_BINARY_NAME=raiden
CLI_VERSION=v1.0.0
BUILD_DIR=build

build: check-build build-linux build-windows build-macos

check-build:
	@mkdir -p $(BUILD_DIR)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_linux_$(CLI_VERSION) ./cmd/raiden/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_windows_$(CLI_VERSION).exe ./cmd/raiden/main.go

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_macos_$(CLI_VERSION) ./cmd/raiden/main.go

clean:
	go clean
	rm -f $(BINARY_NAME)_*

.PHONY: build build-linux build-windows build-macos clean
