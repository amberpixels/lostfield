# Variables
export PATH := env("PATH") + ":" + `go env GOPATH` + "/bin"
binary_name := "lostfield"
install_path := `go env GOPATH` + "/bin"

# Default target
default: tidy

# Tidy: format and vet the code
tidy:
    @go fmt $(go list ./...)
    @go vet $(go list ./...)
    @go mod tidy

# Install golangci-lint only if it's not already installed
lint-install:
    @if ! which golangci-lint > /dev/null 2>&1; then \
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
    fi

# Lint the code using golangci-lint
lint: lint-install
    $(which golangci-lint) run

# Run tests
test:
    @go test -v $(go list ./...)

# Install the linter binary to GOPATH/bin
install:
    @echo "Installing {{ binary_name }} to {{ install_path }}..."
    @go install ./cmd/lostfield
    @echo "✓ {{ binary_name }} installed successfully"
    @echo "You can now run 'go vet -vettool=$(which {{ binary_name }}) ./...' in any Go project"

# Run the linter on a specified path (usage: just run [TARGET=path])
run TARGET=".":
    @cd {{ TARGET }} && go vet -vettool=$(which {{ binary_name }}) ./... || true
