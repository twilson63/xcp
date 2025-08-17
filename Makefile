.PHONY: build test lint clean build-all build-linux build-darwin build-windows clean-dist checksums

build:
	go build -o bin/xcp ./cmd/xcp

run:
	go run ./cmd/xcp/main.go

test:
	go test ./...

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -rf bin/

# Cross-compilation targets for distribution
build-all: build-linux build-darwin build-windows

# Linux builds
build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-linux-amd64 ./cmd/xcp
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/xcp-linux-arm64 ./cmd/xcp

# macOS builds
build-darwin:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-darwin-amd64 ./cmd/xcp
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/xcp-darwin-arm64 ./cmd/xcp

# Windows builds
build-windows:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-windows-amd64.exe ./cmd/xcp

# Generate checksums for all binaries
checksums: build-all
	cd dist && (sha256sum * 2>/dev/null || shasum -a 256 *) > checksums.txt

# Clean distribution directory
clean-dist:
	rm -rf dist/