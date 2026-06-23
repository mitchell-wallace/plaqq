# Local dev version: the VERSION file's release as a base, marked as a dev build
# with the short commit hash (e.g. 0.3.0-dev+1a06be9, plus .dirty when the tree
# has uncommitted changes). Release binaries are versioned by goreleaser instead.
version := shell('base="$(tr -d "[:space:]" < VERSION 2>/dev/null || echo 0.0.0)"; hash="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"; dirty=""; [ -n "$(git status --porcelain 2>/dev/null)" ] && dirty=".dirty"; printf "%s-dev+%s%s" "$base" "$hash" "$dirty"')
gopath := shell("go env GOPATH")

default: build

# Build the plaqq binary
build:
	go build -ldflags "-X main.version={{version}}" -o bin/plaqq ./cmd/plaqq

# Run all tests
test:
	go test ./...

# Render every font to inspectable artifacts (PNG + .ansi + manifest.json) under
# artifacts/visual/. The PNGs paint each terminal cell's exact sub-cell geometry,
# so they show the true glyph shapes for design review without depending on a
# system font. Inspect artifacts/visual/<font>.png; use the .ansi files for
# exact cell-level debugging.
visual *args:
	go run ./cmd/fontgallery -out artifacts/visual {{args}}

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

# Install the plaqq binary locally to ~/.local/bin.
# Copy to a temp file in the same dir, then atomically rename over the target so
# the install succeeds even while an old plaqq is still running (avoids the
# "text file busy" error from overwriting an in-use binary in place).
install: build
	mkdir -p ~/.local/bin
	cp bin/plaqq ~/.local/bin/.plaqq.new
	chmod +x ~/.local/bin/.plaqq.new
	mv -f ~/.local/bin/.plaqq.new ~/.local/bin/plaqq


