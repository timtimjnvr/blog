package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/substitution"
)

func TestNew(t *testing.T) {
	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry)

	if gen == nil {
		t.Fatal("New returned nil")
	}
	if gen.registry != registry {
		t.Error("registry not set correctly")
	}
	if gen.converter == nil {
		t.Error("converter is nil")
	}
	if gen.template == "" {
		t.Error("template is empty")
	}
}

func TestGenerator_WithTemplate(t *testing.T) {
	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry)

	customTemplate := "<html>{{content}}</html>"
	gen.WithTemplate(customTemplate)

	if gen.template != customTemplate {
		t.Errorf("WithTemplate() did not set template, got %q", gen.template)
	}
}

func TestGenerator_WithTemplate_Chaining(t *testing.T) {
	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("<html></html>")

	if gen == nil {
		t.Error("WithTemplate() should return generator for chaining")
	}
}

func TestGenerator_Generate(t *testing.T) {
	// Create temporary directories
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create test markdown file
	testMd := "# Test Title\n\nTest content."
	err := os.WriteFile(filepath.Join(contentDir, "test.md"), []byte(testMd), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Setup generator with substitutions
	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.TitleSubstituter{})
	registry.Register(&substitution.ContentSubstituter{})

	gen := New(registry).WithTemplate("<title>{{title}}</title>\n<body>{{content}}</body>")

	// Generate
	err = gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify output
	outputPath := filepath.Join(buildDir, "test.html")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if !strings.Contains(string(content), "<title>Test Title</title>") {
		t.Errorf("output should contain title, got %s", content)
	}
	if !strings.Contains(string(content), "<h1>Test Title</h1>") {
		t.Errorf("output should contain h1, got %s", content)
	}
}

func TestGenerator_Generate_CreatesDirectories(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create nested content
	postsDir := filepath.Join(contentDir, "posts")
	os.MkdirAll(postsDir, 0755)
	err := os.WriteFile(filepath.Join(postsDir, "article.md"), []byte("# Article"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err = gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify nested output directory was created
	outputPath := filepath.Join(buildDir, "post", "article.html")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("expected output file at %s", outputPath)
	}
}

func TestGenerator_Generate_SkipsNonMarkdown(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create markdown and non-markdown files
	os.WriteFile(filepath.Join(contentDir, "page.md"), []byte("# Page"), 0644)
	os.WriteFile(filepath.Join(contentDir, "image.png"), []byte("fake image"), 0644)
	os.WriteFile(filepath.Join(contentDir, "style.css"), []byte("body {}"), 0644)

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Only markdown should be converted
	if _, err := os.Stat(filepath.Join(buildDir, "page.html")); os.IsNotExist(err) {
		t.Error("page.html should exist")
	}
	if _, err := os.Stat(filepath.Join(buildDir, "image.html")); !os.IsNotExist(err) {
		t.Error("image.html should not exist")
	}
	if _, err := os.Stat(filepath.Join(buildDir, "style.html")); !os.IsNotExist(err) {
		t.Error("style.html should not exist")
	}
}

func TestGenerator_Generate_EmptyDirectory(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() should not error on empty directory, got %v", err)
	}
}

func TestGenerator_Generate_InvalidContentDir(t *testing.T) {
	buildDir := t.TempDir()

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate("/nonexistent/path", buildDir)
	if err == nil {
		t.Error("Generate() should error for nonexistent content directory")
	}
}

func TestGenerator_Generate_HomePageRouting(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// home.md should become index.html
	os.WriteFile(filepath.Join(contentDir, "home.md"), []byte("# Home"), 0644)

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(buildDir, "index.html")); os.IsNotExist(err) {
		t.Error("home.md should become index.html")
	}
}
