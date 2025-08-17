package testing

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MockFileSystem provides a mock file system for testing
type MockFileSystem struct {
	Files     map[string][]byte
	Dirs      map[string]bool
	mu        sync.RWMutex
	StdoutBuf *bytes.Buffer
	StderrBuf *bytes.Buffer
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:     make(map[string][]byte),
		Dirs:      make(map[string]bool),
		StdoutBuf: new(bytes.Buffer),
		StderrBuf: new(bytes.Buffer),
	}
}

// Stdout returns a writer for stdout
func (m *MockFileSystem) Stdout() io.Writer {
	return m.StdoutBuf
}

// Stderr returns a writer for stderr
func (m *MockFileSystem) Stderr() io.Writer {
	return m.StderrBuf
}

// ReadFile mocks reading a file
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.Files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// WriteFile mocks writing a file
func (m *MockFileSystem) WriteFile(path string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	m.Dirs[dir] = true

	m.Files[path] = data
	return nil
}

// MkdirAll mocks creating directories
func (m *MockFileSystem) MkdirAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Dirs[path] = true
	return nil
}

// Stat mocks getting file info
func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if it's a directory
	if _, isDir := m.Dirs[path]; isDir {
		return &mockFileInfo{name: filepath.Base(path), isDir: true}, nil
	}

	// Check if it's a file
	if _, isFile := m.Files[path]; isFile {
		return &mockFileInfo{name: filepath.Base(path), isDir: false}, nil
	}

	return nil, os.ErrNotExist
}

// FileExists checks if a file exists
func (m *MockFileSystem) FileExists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.Files[path]
	return exists
}

// DirExists checks if a directory exists
func (m *MockFileSystem) DirExists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.Dirs[path]
	return exists
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime int64
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Unix(m.mtime, 0) }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
