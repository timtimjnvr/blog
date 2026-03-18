package article

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArticlePrint(t *testing.T) {
	a := Article{name: "Hello", filePath: "hello.md"}
	if got := a.Print(); got != "- [Hello](hello.md)" {
		t.Errorf("Print() = %q, want %q", got, "- [Hello](hello.md)")
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty content", input: "", want: ""},
		{name: "no heading", input: "some text\nno heading here", want: ""},
		{name: "hash without space", input: "#notATitle", want: ""},
		{name: "valid h1", input: "# My Title\nsome content", want: "My Title"},
		{name: "multiple h1s returns first", input: "# First\n# Second", want: "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTitle([]byte(tt.input))
			if got != tt.want {
				t.Errorf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListPrinters(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string // relative path -> content
		indexFile string            // relative path of the index file
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "empty directory",
			files:     map[string]string{"index.md": "# Index"},
			indexFile: "index.md",
			wantNames: []string{},
		},
		{
			name: "two articles",
			files: map[string]string{
				"index.md":  "# Index",
				"hello.md":  "# Hello World\nsome content",
				"second.md": "# Second Post\nmore content",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello World", "Second Post"},
		},
		{
			name: "skips file without h1",
			files: map[string]string{
				"index.md":   "# Index",
				"notitle.md": "no heading here",
				"hello.md":   "# Hello\ncontent",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello"},
		},
		{
			name: "skips subdirectory",
			files: map[string]string{
				"index.md":     "# Index",
				"sub/child.md": "# Child",
				"hello.md":     "# Hello\ncontent",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for relPath, content := range tt.files {
				fullPath := filepath.Join(dir, relPath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}

			lister := NewPageArticlesLister(filepath.Join(dir, tt.indexFile))
			articles, err := lister.ListPrinters()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(articles) != len(tt.wantNames) {
				t.Fatalf("got %d articles, want %d", len(articles), len(tt.wantNames))
			}

			gotNames := make(map[string]bool)
			for _, a := range articles {
				gotNames[a.name] = true
			}
			for _, name := range tt.wantNames {
				if !gotNames[name] {
					t.Errorf("missing article with name %q", name)
				}
			}
		})
	}
}
