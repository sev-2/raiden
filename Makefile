CLI_BINARY_NAME=raiden
CLI_VERSION=v1.0.0
BUILD_DIR=build

build: check-build build-linux build-windows build-macos

build-arm64: check-build build-linux-arm64 build-macos-arm64

check-build:
	@mkdir -p $(BUILD_DIR)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_linux_$(CLI_VERSION) ./cmd/raiden/main.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_linux_arm64_$(CLI_VERSION) ./cmd/cli/raiden.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_windows_$(CLI_VERSION).exe ./cmd/raiden/main.go

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_macos_$(CLI_VERSION) ./cmd/raiden/main.go

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME)_macos_arm64_$(CLI_VERSION) ./cmd/cli/raiden.go

clean:
	go clean
	rm -f $(BINARY_NAME)_*

.PHONY: build build-linux build-windows build-macos clean
