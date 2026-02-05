.PHONY: build install clean test lint

VERSION ?= dev
LDFLAGS = -ldflags "-X github.com/josinSbazin/gf/internal/version.Version=$(VERSION)"

# Build for current platform
build:
	go build $(LDFLAGS) -o gf .

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) .

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/gf-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/gf-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/gf-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/gf-windows-amd64.exe .

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f gf gf.exe
	rm -rf dist/

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy
