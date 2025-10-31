# Variables
export PATH := $(PATH):$(shell go env GOPATH)/bin
GOLANGCI_LINT := $(shell which golangci-lint)
BINARY_NAME := lostfield
INSTALL_PATH := $(shell go env GOPATH)/bin

# Default target
all: tidy

# Tidy: format and vet the code
tidy:
	@go fmt $$(go list ./...)
	@go vet $$(go list ./...)
	@go mod tidy

# Install golangci-lint only if it's not already installed
lint-install:
	@if ! [ -x "$(GOLANGCI_LINT)" ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Lint the code using golangci-lint
# todo reuse var if possible
lint: lint-install
	$(shell which golangci-lint) run

test:
	@go test -v $$(go list ./...)

# Install the linter binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@go install ./cmd/lostfield
	@echo "✓ $(BINARY_NAME) installed successfully"
	@echo "You can now run 'go vet -vettool=\$$(which $(BINARY_NAME)) ./...' in any Go project"

# Run the linter on a specified path (usage: make run [TARGET=path])
run:
	@TARGET=$${TARGET:-.}; \
	cd $$TARGET && go vet -vettool=$$(which $(BINARY_NAME)) ./... || true

# Phony targets
.PHONY: all tidy lint-install lint test install run
