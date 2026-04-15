package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MemoryFileSystem implements FileSystem in-memory for testing
type MemoryFileSystem struct {
	files        map[string][]byte
	dirs         map[string]bool
	MkdirAllErr  error
	WriteFileErr error
	StatErr      error
}

// memFileInfo is a minimal os.FileInfo implementation for MemoryFileSystem
type memFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (fi memFileInfo) Name() string      { return fi.name }
func (fi memFileInfo) Size() int64       { return fi.size }
func (fi memFileInfo) Mode() os.FileMode { return 0644 }
func (fi memFileInfo) ModTime() time.Time { return time.Time{} }
func (fi memFileInfo) IsDir() bool       { return fi.isDir }
func (fi memFileInfo) Sys() interface{}  { return nil }

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

func (m *MemoryFileSystem) Stat(path string) (os.FileInfo, error) {
	if m.StatErr != nil {
		return nil, m.StatErr
	}
	if _, ok := m.files[path]; ok {
		return memFileInfo{name: filepath.Base(path), size: int64(len(m.files[path])), isDir: false}, nil
	}
	if m.dirs[path] {
		return memFileInfo{name: filepath.Base(path), isDir: true}, nil
	}
	return nil, fmt.Errorf("stat %s: no such file or directory", path)
}

func (m *MemoryFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	// Collect all known paths under root
	paths := make([]string, 0)
	paths = append(paths, root)
	for path := range m.files {
		if path == root || isUnder(path, root) {
			paths = append(paths, path)
		}
	}
	for path := range m.dirs {
		if path == root || isUnder(path, root) {
			paths = append(paths, path)
		}
	}
	// Deduplicate and sort for deterministic order
	seen := make(map[string]bool)
	unique := paths[:0]
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}
	sort.Strings(unique)

	for _, path := range unique {
		info, err := m.Stat(path)
		if err != nil {
			if walkErr := fn(path, nil, err); walkErr != nil {
				return walkErr
			}
			continue
		}
		if err := fn(path, info, nil); err != nil {
			return err
		}
	}
	return nil
}

// isUnder reports whether path is under root (i.e., root is a prefix component)
func isUnder(path, root string) bool {
	if root == "." || root == "" {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && len(rel) >= 2 || (len(rel) > 0 && rel[0] != '.')
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
