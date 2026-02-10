package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMemoryFileSystem(t *testing.T) {
	fs := NewMemoryFileSystem()
	if fs == nil {
		t.Fatal("NewMemoryFileSystem() returned nil")
	}
}

func TestMemoryFileSystem_ReadFile(t *testing.T) {
	t.Run("reads existing file", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		fs.AddFile("/test.md", []byte("hello"))

		data, err := fs.ReadFile("/test.md")
		if err != nil {
			t.Fatalf("ReadFile() unexpected error: %v", err)
		}
		if string(data) != "hello" {
			t.Errorf("ReadFile() = %q, want %q", string(data), "hello")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		_, err := fs.ReadFile("/missing.md")
		if err == nil {
			t.Error("ReadFile() expected error for missing file, got nil")
		}
	})
}

func TestMemoryFileSystem_WriteFile(t *testing.T) {
	fs := NewMemoryFileSystem()

	err := fs.WriteFile("/output.html", []byte("<html></html>"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	data, ok := fs.GetFile("/output.html")
	if !ok {
		t.Fatal("GetFile() file not found after WriteFile")
	}
	if string(data) != "<html></html>" {
		t.Errorf("GetFile() = %q, want %q", string(data), "<html></html>")
	}
}

func TestMemoryFileSystem_WriteFile_Overwrites(t *testing.T) {
	fs := NewMemoryFileSystem()
	_ = fs.WriteFile("/file.txt", []byte("first"), 0644)
	_ = fs.WriteFile("/file.txt", []byte("second"), 0644)

	data, ok := fs.GetFile("/file.txt")
	if !ok {
		t.Fatal("GetFile() file not found")
	}
	if string(data) != "second" {
		t.Errorf("WriteFile() should overwrite, got %q", string(data))
	}
}

func TestMemoryFileSystem_MkdirAll(t *testing.T) {
	fs := NewMemoryFileSystem()
	err := fs.MkdirAll("/build/posts", 0755)
	if err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}
}

func TestMemoryFileSystem_AddFile(t *testing.T) {
	fs := NewMemoryFileSystem()
	fs.AddFile("/content/post.md", []byte("# Title"))

	data, err := fs.ReadFile("/content/post.md")
	if err != nil {
		t.Fatalf("ReadFile() after AddFile() unexpected error: %v", err)
	}
	if string(data) != "# Title" {
		t.Errorf("ReadFile() = %q, want %q", string(data), "# Title")
	}
}

func TestMemoryFileSystem_GetFile(t *testing.T) {
	t.Run("returns data and true for existing file", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		fs.AddFile("/test.txt", []byte("content"))

		data, ok := fs.GetFile("/test.txt")
		if !ok {
			t.Error("GetFile() returned false for existing file")
		}
		if string(data) != "content" {
			t.Errorf("GetFile() = %q, want %q", string(data), "content")
		}
	})

	t.Run("returns false for missing file", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		_, ok := fs.GetFile("/missing.txt")
		if ok {
			t.Error("GetFile() returned true for missing file")
		}
	})
}

func TestMemoryFileSystem_ImplementsFileSystem(t *testing.T) {
	// Compile-time check that MemoryFileSystem implements FileSystem
	var _ FileSystem = (*MemoryFileSystem)(nil)
}

func TestOSFileSystem_ImplementsFileSystem(t *testing.T) {
	// Compile-time check that OSFileSystem implements FileSystem
	var _ FileSystem = OSFileSystem{}
}

func TestNewOSFileSystem(t *testing.T) {
	fs := NewOSFileSystem()
	if fs == nil {
		t.Fatal("NewOSFileSystem() returned nil")
	}
}

func TestOSFileSystem_RoundTrip(t *testing.T) {
	fs := NewOSFileSystem()
	dir := t.TempDir()

	subDir := filepath.Join(dir, "sub", "dir")
	err := fs.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	info, err := os.Stat(subDir)
	if err != nil {
		t.Fatalf("created directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll() should create a directory")
	}

	filePath := filepath.Join(subDir, "test.txt")
	err = fs.WriteFile(filePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	data, err := fs.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("ReadFile() = %q, want %q", string(data), "hello")
	}
}

func TestOSFileSystem_ReadFile_NotFound(t *testing.T) {
	fs := NewOSFileSystem()
	_, err := fs.ReadFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("ReadFile() expected error for nonexistent file, got nil")
	}
}
