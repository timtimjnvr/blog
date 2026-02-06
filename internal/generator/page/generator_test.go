package page

import (
	"fmt"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/generator/page/filesystem"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/page/substitution"
	"github.com/timtimjnvr/blog/internal/generator/page/validation"
)

func newTestGenerator(t *testing.T, markdownPath, buildDir, sectionName string, fs *filesystem.MemoryFileSystem) *Generator {
	t.Helper()
	config := styling.Config{
		Elements: make(map[string]string),
		Contexts: make(map[string]map[string]string),
	}
	subs := substitution.NewRegistry(nil, sectionName)
	vals := validation.NewRegistry(nil)
	return NewGenerator(markdownPath, buildDir, sectionName, config, fs, subs, vals)
}

func TestNewGenerator(t *testing.T) {
	t.Run("sets output path without section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/index.md", "/build", "", fs)

		if g.htmlOutputPath != "/build/index.html" {
			t.Errorf("htmlOutputPath = %q, want %q", g.htmlOutputPath, "/build/index.html")
		}
	})

	t.Run("sets output path with section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/posts/hello.md", "/build", "posts", fs)

		if g.htmlOutputPath != "/build/posts/hello.html" {
			t.Errorf("htmlOutputPath = %q, want %q", g.htmlOutputPath, "/build/posts/hello.html")
		}
	})

	t.Run("stores all fields", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/page.md", "/out", "blog", fs)

		if g.markdownPath != "/content/page.md" {
			t.Errorf("markdownPath = %q, want %q", g.markdownPath, "/content/page.md")
		}
		if g.buildDir != "/out" {
			t.Errorf("buildDir = %q, want %q", g.buildDir, "/out")
		}
		if g.sectionName != "blog" {
			t.Errorf("sectionName = %q, want %q", g.sectionName, "blog")
		}
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Run("generates html from markdown", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/index.md", []byte("# Hello World\n\nSome content here."))

		g := newTestGenerator(t, "/content/index.md", "/build", "", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/index.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, "Hello World") {
			t.Errorf("output should contain title, got %q", html)
		}
		if !strings.Contains(html, "Some content here.") {
			t.Errorf("output should contain content, got %q", html)
		}
		if !strings.Contains(html, "<title>Hello World</title>") {
			t.Errorf("output should have title tag, got %q", html)
		}
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Errorf("output should contain DOCTYPE, got %q", html)
		}
	})

	t.Run("generates html with section name", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/posts/article.md", []byte("# My Article\n\nArticle body."))

		g := newTestGenerator(t, "/content/posts/article.md", "/build", "posts", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		_, ok := fs.GetFile("/build/posts/article.html")
		if !ok {
			t.Fatal("Generate() did not write output file at section path")
		}
	})

	t.Run("returns error when markdown file not found", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/missing.md", "/build", "", fs)

		err := g.Generate()
		if err == nil {
			t.Fatal("Generate() expected error for missing file, got nil")
		}
		if !strings.Contains(err.Error(), "reading") {
			t.Errorf("error should mention reading, got %q", err.Error())
		}
	})

	t.Run("stores html content bytes after generation", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/page.md", []byte("# Title\n\nBody text."))

		g := newTestGenerator(t, "/content/page.md", "/build", "", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		if len(g.htmlContentBytes) == 0 {
			t.Error("htmlContentBytes should be populated after Generate()")
		}
	})
}

func TestGenerator_Generate_MkdirAllError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/page.md", []byte("# Title\n\nContent."))
	fs.MkdirAllErr = fmt.Errorf("permission denied")

	g := newTestGenerator(t, "/content/page.md", "/build", "", fs)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when MkdirAll fails, got nil")
	}
	if !strings.Contains(err.Error(), "output directory") {
		t.Errorf("error should mention output directory, got %q", err.Error())
	}
}

func TestGenerator_Generate_WriteFileError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/page.md", []byte("# Title\n\nContent."))
	fs.WriteFileErr = fmt.Errorf("disk full")

	g := newTestGenerator(t, "/content/page.md", "/build", "", fs)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when WriteFile fails, got nil")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Errorf("error should mention write, got %q", err.Error())
	}
}

func TestGenerator_Generate_SubstitutionError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	// Markdown without h1 will cause title substitution to fail
	fs.AddFile("/content/notitle.md", []byte("No heading here, just text."))

	config := styling.Config{
		Elements: make(map[string]string),
		Contexts: make(map[string]map[string]string),
	}
	subs := substitution.NewRegistry(nil, "")
	vals := validation.NewRegistry(nil)
	g := NewGenerator("/content/notitle.md", "/build", "", config, fs, subs, vals)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when title substitution fails, got nil")
	}
	if !strings.Contains(err.Error(), "template") {
		t.Errorf("error should mention template projection, got %q", err.Error())
	}
}

func TestGenerator_Generate_NavigationBar(t *testing.T) {
	t.Run("generates navigation with sections from root", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/index.md", []byte("# Home\n\nWelcome."))

		config := styling.Config{
			Elements: make(map[string]string),
			Contexts: make(map[string]map[string]string),
		}
		subs := substitution.NewRegistry([]string{"posts", "about"}, "")
		vals := validation.NewRegistry(nil)
		g := NewGenerator("/content/index.md", "/build", "", config, fs, subs, vals)

		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/index.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, "<nav") {
			t.Error("output should contain nav element")
		}
		if !strings.Contains(html, `href="index.html"`) {
			t.Errorf("output should contain home link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="posts/index.html"`) {
			t.Errorf("output should contain posts link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="about/index.html"`) {
			t.Errorf("output should contain about link, got:\n%s", html)
		}
	})

	t.Run("generates navigation with relative paths from section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/posts/hello.md", []byte("# Hello\n\nPost content."))

		config := styling.Config{
			Elements: make(map[string]string),
			Contexts: make(map[string]map[string]string),
		}
		subs := substitution.NewRegistry([]string{"posts", "about"}, "posts")
		vals := validation.NewRegistry(nil)
		g := NewGenerator("/content/posts/hello.md", "/build", "posts", config, fs, subs, vals)

		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/posts/hello.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, `href="../index.html"`) {
			t.Errorf("output should contain relative home link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="../posts/index.html"`) {
			t.Errorf("output should contain relative posts link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="../about/index.html"`) {
			t.Errorf("output should contain relative about link, got:\n%s", html)
		}
	})
}

func TestGenerator_Validate(t *testing.T) {
	t.Run("validate returns nil with empty registry", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/page.md", []byte("# Page\n\nContent."))

		g := newTestGenerator(t, "/content/page.md", "/build", "", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		err = g.Validate()
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})
}
