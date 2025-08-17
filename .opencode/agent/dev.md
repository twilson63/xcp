# XCP Development Agent

## Overview
This agent assists with development tasks for the XCP (eXtreme Copy) project, a Go-based CLI tool for downloading GitHub repositories and directories.

## Key Commands
- **Build**: `go build ./cmd/xcp`
- **Run**: `go run ./cmd/xcp/main.go`
- **Test**: `go test ./...`
- **Lint**: `go vet ./...`
- **Format**: `go fmt ./...`

## Project Structure
- `cmd/xcp/` - Main application entry point
- `internal/cli/` - CLI interface and command handling
- `internal/downloader/` - Download logic and file operations
- `internal/github/` - GitHub API client and URL parsing
- `internal/testing/` - Test utilities and mocks

## Development Guidelines
- Follow Go standard formatting (`go fmt`)
- Check all errors; avoid panic in production
- Write unit tests for business logic
- Use strong typing; avoid `interface{}`
- Document exported functions and types
- Keep functions small and focused
- Group imports: stdlib, third-party, local

## Common Tasks
- Add new CLI commands in `internal/cli/`
- Extend download functionality in `internal/downloader/`
- Enhance GitHub integration in `internal/github/`
- Write tests with mocks from `internal/testing/`

## Testing Strategy
- Unit tests for all packages
- Integration tests for CLI commands
- Mock GitHub API responses
- Test error conditions and edge cases