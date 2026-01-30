package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/styling"
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
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatalf("failed to create posts dir: %v", err)
	}
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
	outputPath := filepath.Join(buildDir, "posts", "article.html")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("expected output file at %s", outputPath)
	}
}

func TestGenerator_Generate_SkipsNonMarkdown(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create markdown and non-markdown files
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte("# Page"), 0644); err != nil {
		t.Fatalf("failed to create page.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "image.png"), []byte("fake image"), 0644); err != nil {
		t.Fatalf("failed to create image.png: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "style.css"), []byte("body {}"), 0644); err != nil {
		t.Fatalf("failed to create style.css: %v", err)
	}

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

func TestGenerator_Generate_IndexPageRouting(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// index.md should become index.html
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte("# Home"), 0644); err != nil {
		t.Fatalf("failed to create index.md: %v", err)
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(buildDir, "index.html")); os.IsNotExist(err) {
		t.Error("index.md should become index.html")
	}
}

// Integration tests for styling system

func TestGenerator_Integration_StyleConfigApplied(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create markdown with various elements
	markdown := `# Main Title

This is a paragraph with a [link](https://example.com).

![image](test.png)

> A blockquote

- List item 1
- List item 2
`
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte(markdown), 0644); err != nil {
		t.Fatalf("failed to create markdown: %v", err)
	}

	// Create style config
	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1":   "text-4xl font-bold",
			"paragraph":  "text-base leading-relaxed",
			"link":       "text-blue-600 hover:underline",
			"image":      "rounded-lg shadow-md",
			"blockquote": "border-l-4 italic",
			"list":       "list-disc pl-6",
		},
		Contexts: make(map[string]map[string]string),
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.ContentSubstituter{})

	gen := New(registry).
		WithTemplate("{{content}}").
		WithStyleConfig(styleConfig)

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Read generated HTML
	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	// Verify all style classes are applied
	checks := []struct {
		desc     string
		contains string
	}{
		{"heading has class", `class="text-4xl font-bold"`},
		{"link has class", `class="text-blue-600 hover:underline"`},
		{"image has class", `class="rounded-lg shadow-md"`},
		{"blockquote has class", `class="border-l-4 italic"`},
		{"list has class", `class="list-disc pl-6"`},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.contains) {
			t.Errorf("%s: expected %q in output, got:\n%s", check.desc, check.contains, html)
		}
	}
}

func TestGenerator_Integration_InlineAttributesOverrideConfig(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Markdown with inline attribute that should override config
	markdown := `# Title {.custom-inline-class}

## Regular Heading
`
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte(markdown), 0644); err != nil {
		t.Fatalf("failed to create markdown: %v", err)
	}

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1": "config-h1-class",
			"heading2": "config-h2-class",
		},
		Contexts: make(map[string]map[string]string),
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.ContentSubstituter{})

	gen := New(registry).
		WithTemplate("{{content}}").
		WithStyleConfig(styleConfig)

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	// H1 should have inline class (override)
	if !strings.Contains(html, `class="custom-inline-class"`) {
		t.Errorf("H1 should have inline class, got:\n%s", html)
	}
	if strings.Contains(html, "config-h1-class") {
		t.Errorf("H1 should NOT have config class when inline exists, got:\n%s", html)
	}

	// H2 should have config class (no inline override)
	if !strings.Contains(html, `class="config-h2-class"`) {
		t.Errorf("H2 should have config class, got:\n%s", html)
	}
}

func TestGenerator_Integration_ContextSpecificStyling(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Create posts directory
	postsDir := filepath.Join(contentDir, "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatalf("failed to create posts dir: %v", err)
	}

	// Create a regular page and a post
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte("# Regular Page"), 0644); err != nil {
		t.Fatalf("failed to create page.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(postsDir, "article.md"), []byte("# Post Title"), 0644); err != nil {
		t.Fatalf("failed to create article.md: %v", err)
	}

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1": "global-heading-style",
		},
		Contexts: map[string]map[string]string{
			"post": {
				"heading1": "post-heading-style",
			},
		},
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.ContentSubstituter{})

	gen := New(registry).
		WithTemplate("{{content}}").
		WithStyleConfig(styleConfig)

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check regular page uses global style
	pageContent, _ := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if !strings.Contains(string(pageContent), `class="global-heading-style"`) {
		t.Errorf("Regular page should have global style, got:\n%s", pageContent)
	}

	// Check post uses context-specific style
	postContent, _ := os.ReadFile(filepath.Join(buildDir, "posts", "article.html"))
	if !strings.Contains(string(postContent), `class="post-heading-style"`) {
		t.Errorf("Post should have post-specific style, got:\n%s", postContent)
	}
}

func TestGenerator_Integration_NoStyleConfig(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Markdown with inline attributes should still work without config
	markdown := `# Title {.my-class #my-id}
`
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte(markdown), 0644); err != nil {
		t.Fatalf("failed to create markdown: %v", err)
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.ContentSubstituter{})

	// No style config
	gen := New(registry).WithTemplate("{{content}}")

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	// Inline attributes should still work
	if !strings.Contains(html, `class="my-class"`) {
		t.Errorf("should have inline class, got:\n%s", html)
	}
	if !strings.Contains(html, `id="my-id"`) {
		t.Errorf("should have inline id, got:\n%s", html)
	}
}

func TestGenerator_Integration_ComplexMarkdownWithStyling(t *testing.T) {
	contentDir := t.TempDir()
	buildDir := t.TempDir()

	// Complex markdown with mixed styling
	markdown := `# Main Title {.hero-title}

Regular paragraph here.

## Section {#section-1}

[External link](https://example.com)

### Code Example

` + "```go\nfunc main() {}\n```" + `

> Important quote

1. First item
2. Second item
`
	if err := os.WriteFile(filepath.Join(contentDir, "page.md"), []byte(markdown), 0644); err != nil {
		t.Fatalf("failed to create markdown: %v", err)
	}

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading2":   "section-heading",
			"heading3":   "subsection-heading",
			"paragraph":  "prose-paragraph",
			"link":       "external-link",
			"blockquote": "quote-style",
			"list":       "numbered-list",
		},
		Contexts: make(map[string]map[string]string),
	}

	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.ContentSubstituter{})

	gen := New(registry).
		WithTemplate("{{content}}").
		WithStyleConfig(styleConfig)

	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	// Verify mixed inline and config styling
	checks := []struct {
		desc     string
		contains string
	}{
		{"H1 has inline class", `class="hero-title"`},
		{"H2 has config class and inline id", `id="section-1"`},
		{"H2 has config class", `class="section-heading"`},
		{"H3 has config class", `class="subsection-heading"`},
		{"link has config class", `class="external-link"`},
		{"blockquote has config class", `class="quote-style"`},
		{"list has config class", `class="numbered-list"`},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.contains) {
			t.Errorf("%s: expected %q in output, got:\n%s", check.desc, check.contains, html)
		}
	}
}
