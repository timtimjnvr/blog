package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name          string
		srcFiles      map[string]string // relative path -> content
		filter        func(string) bool
		expectedFiles []string // relative paths expected in destDir
		wantErr       bool
	}{
		{
			name: "copy all files without filter",
			srcFiles: map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			},
			filter:        nil,
			expectedFiles: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "copy with extension filter",
			srcFiles: map[string]string{
				"script.js":  "console.log('hello')",
				"style.css":  "body {}",
				"app.js":     "const x = 1",
				"readme.txt": "readme",
			},
			filter: func(path string) bool {
				return strings.HasSuffix(path, ".js")
			},
			expectedFiles: []string{"script.js", "app.js"},
		},
		{
			name: "preserve directory structure",
			srcFiles: map[string]string{
				"root.txt":          "root",
				"sub/nested.txt":    "nested",
				"sub/deep/file.txt": "deep",
			},
			filter:        nil,
			expectedFiles: []string{"root.txt", "sub/nested.txt", "sub/deep/file.txt"},
		},
		{
			name:          "empty source directory",
			srcFiles:      map[string]string{},
			filter:        nil,
			expectedFiles: []string{},
		},
		{
			name: "filter excludes all files",
			srcFiles: map[string]string{
				"file.txt": "content",
			},
			filter: func(path string) bool {
				return false
			},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcDir := filepath.Join(tmpDir, "src")
			destDir := filepath.Join(tmpDir, "dest")

			if err := os.MkdirAll(srcDir, 0755); err != nil {
				t.Fatal(err)
			}

			for relPath, content := range tt.srcFiles {
				fullPath := filepath.Join(srcDir, relPath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			err := copyDir(srcDir, destDir, tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("copyDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			for _, expectedFile := range tt.expectedFiles {
				destPath := filepath.Join(destDir, expectedFile)
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Errorf("expected file %s not found in dest", expectedFile)
					continue
				}

				srcPath := filepath.Join(srcDir, expectedFile)
				srcContent, _ := os.ReadFile(srcPath)
				destContent, _ := os.ReadFile(destPath)

				if string(srcContent) != string(destContent) {
					t.Errorf("content mismatch for %s: got %q, want %q", expectedFile, destContent, srcContent)
				}
			}

			// Verify no extra files were copied
			var copiedFiles []string
			_ = filepath.WalkDir(destDir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				relPath, _ := filepath.Rel(destDir, path)
				copiedFiles = append(copiedFiles, relPath)
				return nil
			})

			if len(copiedFiles) != len(tt.expectedFiles) {
				t.Errorf("got %d files, want %d\ngot: %v\nwant: %v",
					len(copiedFiles), len(tt.expectedFiles), copiedFiles, tt.expectedFiles)
			}
		})
	}
}

func TestCopyDir_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "nonexistent")
	destDir := filepath.Join(tmpDir, "dest")

	err := copyDir(srcDir, destDir, nil)
	if err == nil {
		t.Error("expected error for non-existent source directory")
	}
}
