.PHONY: build test lint clean

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