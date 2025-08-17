# Agent Guidelines for XCP Project

## Build & Test Commands
- Build: `go build ./cmd/xcp`
- Run: `go run ./cmd/xcp/main.go`
- Test: `go test ./...`
- Test single package: `go test ./path/to/package`
- Test single function: `go test -run TestFunctionName ./path/to/package`
- Lint: `go vet ./...`
- Format: `go fmt ./...`

## Code Style Guidelines
- Imports: Group standard library imports first, then third-party, then local packages
- Formatting: Follow Go standard formatting (run `go fmt` before committing)
- Types: Use strong typing; avoid interface{} where possible
- Error handling: Check all errors; don't use panic in production code
- Naming: Use CamelCase for exported items, camelCase for non-exported
- Functions: Keep functions small and focused on a single responsibility
- Comments: Document all exported functions, types, and constants
- File structure: One package per directory; package name matches directory
- Testing: Write unit tests for all business logic
- Error messages: Should be specific and actionable