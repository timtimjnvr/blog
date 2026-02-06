package filesystem

import (
	"fmt"
	"os"
)

// MemoryFileSystem implements FileSystem in-memory for testing
type MemoryFileSystem struct {
	files        map[string][]byte
	dirs         map[string]bool
	MkdirAllErr  error
	WriteFileErr error
}

func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (m *MemoryFileSystem) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (m *MemoryFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if m.WriteFileErr != nil {
		return m.WriteFileErr
	}
	m.files[path] = data
	return nil
}

func (m *MemoryFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.MkdirAllErr != nil {
		return m.MkdirAllErr
	}
	m.dirs[path] = true
	return nil
}

// AddFile is a test helper to pre-populate files
func (m *MemoryFileSystem) AddFile(path string, content []byte) {
	m.files[path] = content
}

// GetFile is a test helper to retrieve written files
func (m *MemoryFileSystem) GetFile(path string) ([]byte, bool) {
	data, ok := m.files[path]
	return data, ok
}
