version := shell("git describe --tags --always --dirty 2>/dev/null || echo dev")
gopath := shell("go env GOPATH")

default: build

# Build the plaqq binary
build:
	go build -ldflags "-X main.version={{version}}" -o bin/plaqq ./cmd/plaqq

# Run all tests
test:
	go test ./...

# Run the linter
lint:
	which golangci-lint 2>/dev/null || curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b {{gopath}}/bin
	{{gopath}}/bin/golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run the plaqq binary with arguments
run *args:
	go run -ldflags "-X main.version={{version}}" ./cmd/plaqq {{args}}

