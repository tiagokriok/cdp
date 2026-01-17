.PHONY: build install test coverage clean dev help

BINARY_NAME=cdp
INSTALL_PATH=$(HOME)/.local/bin
GO=go
VERSION?=0.1.0

## help: Display this help message
help:
	@echo "CDP - Claude Profile Switcher"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## build: Build the binary
build:
	$(GO) build -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_NAME) cmd/cdp/main.go

## install: Install the binary to ~/.local/bin
install: build
	@mkdir -p $(INSTALL_PATH)
	cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

## test: Run all tests
test:
	$(GO) test -v ./...

## coverage: Generate test coverage report
coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out
	$(GO) clean

## dev: Run the application in development mode
dev:
	$(GO) run cmd/cdp/main.go

## fmt: Format Go code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run all linters
lint: fmt vet
	@echo "Linting complete"
