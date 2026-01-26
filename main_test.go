package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/generator"
	"github.com/timtimjnvr/blog/internal/styling"
	"github.com/timtimjnvr/blog/internal/substitution"
	"github.com/timtimjnvr/blog/internal/validator"
)

// Helper to create a generator with standard configuration
func createGenerator(styleConfig *styling.Config) *generator.Generator {
	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.TitleSubstituter{})
	registry.Register(&substitution.ContentSubstituter{})

	gen := generator.New(registry).
		WithValidator(validator.NewImageValidator())

	if styleConfig != nil {
		gen = gen.WithStyleConfig(styleConfig)
	}

	return gen
}

// Helper to create test content directory structure
func setupTestContent(t *testing.T, files map[string]string) (contentDir, buildDir string) {
	contentDir = t.TempDir()
	buildDir = t.TempDir()

	for path, content := range files {
		fullPath := filepath.Join(contentDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return contentDir, buildDir
}

// Integration tests for the full site generation pipeline

func TestIntegration_BasicSiteGeneration(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"home.md": `# Welcome

This is the homepage.

[Go to posts](posts/index.md)
`,
		"posts/index.md": `# All Posts

- [First Post](first.md)
- [Second Post](second.md)
`,
		"posts/first.md": `# First Post

Content of the first post.

[Back to posts](index.md)
`,
		"posts/second.md": `# Second Post

Content of the second post.
`,
	})

	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify files are generated
	expectedFiles := []string{
		"index.html",
		"posts/index.html",
		"post/first.html",
		"post/second.html",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(buildDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestIntegration_TailwindTemplateApplied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Test Page

Some content here.
`,
	})

	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	checks := []struct {
		desc     string
		contains string
	}{
		{"DOCTYPE", "<!DOCTYPE html>"},
		{"Tailwind CDN", "cdn.tailwindcss.com"},
		{"Typography plugin", "plugins=typography"},
		{"Prose wrapper", `class="prose`},
		{"Article element", "<article"},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.contains) {
			t.Errorf("%s: expected %q in output", check.desc, check.contains)
		}
	}
}

func TestIntegration_LinkConversion(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"home.md": `# Home

[Go to posts](posts/index.md)
`,
		"posts/index.md": `# Posts

[Back to home](../home.md)
[Read article](article.md)
`,
		"posts/article.md": `# Article

[Back to index](index.md)
`,
	})

	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check home.md links
	homeContent, _ := os.ReadFile(filepath.Join(buildDir, "index.html"))
	if !strings.Contains(string(homeContent), `href="posts/index.html"`) {
		t.Errorf("home should link to posts/index.html, got:\n%s", homeContent)
	}

	// Check posts/index.md links
	postsContent, _ := os.ReadFile(filepath.Join(buildDir, "posts/index.html"))
	if !strings.Contains(string(postsContent), `href="../index.html"`) {
		t.Errorf("posts/index should link to ../index.html, got:\n%s", postsContent)
	}
	if !strings.Contains(string(postsContent), `href="../post/article.html"`) {
		t.Errorf("posts/index should link to ../post/article.html, got:\n%s", postsContent)
	}
}

func TestIntegration_StyleConfigApplied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Main Title

A paragraph with a [link](https://example.com).

> A blockquote

- List item
`,
	})

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1":   "custom-h1-class",
			"paragraph":  "custom-p-class",
			"link":       "custom-link-class",
			"blockquote": "custom-quote-class",
			"list":       "custom-list-class",
		},
		Contexts: make(map[string]map[string]string),
	}

	gen := createGenerator(styleConfig)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	checks := []string{
		`class="custom-h1-class"`,
		`class="custom-link-class"`,
		`class="custom-quote-class"`,
		`class="custom-list-class"`,
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("expected %q in output, got:\n%s", check, html)
		}
	}
}

func TestIntegration_InlineAttributesOverrideConfig(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Title With Inline {.inline-class}

## Regular Heading
`,
	})

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1": "config-h1",
			"heading2": "config-h2",
		},
		Contexts: make(map[string]map[string]string),
	}

	gen := createGenerator(styleConfig)
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
	if !strings.Contains(html, `class="inline-class"`) {
		t.Errorf("H1 should have inline class")
	}
	if strings.Contains(html, "config-h1") {
		t.Errorf("H1 should NOT have config class when inline exists")
	}

	// H2 should have config class
	if !strings.Contains(html, `class="config-h2"`) {
		t.Errorf("H2 should have config class, got:\n%s", html)
	}
}

func TestIntegration_ContextSpecificStyling(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md":          "# Regular Page",
		"posts/article.md": "# Post Title",
	})

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1": "global-style",
		},
		Contexts: map[string]map[string]string{
			"post": {
				"heading1": "post-style",
			},
		},
	}

	gen := createGenerator(styleConfig)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Regular page should have global style
	pageContent, _ := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if !strings.Contains(string(pageContent), `class="global-style"`) {
		t.Errorf("regular page should have global-style")
	}

	// Post should have context-specific style
	postContent, _ := os.ReadFile(filepath.Join(buildDir, "post/article.html"))
	if !strings.Contains(string(postContent), `class="post-style"`) {
		t.Errorf("post should have post-style, got:\n%s", postContent)
	}
}

func TestIntegration_InlineAttributesWithoutConfig(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Title {.my-class #my-id}

## Section {data-testid=section-1}
`,
	})

	// No style config
	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `class="my-class"`) {
		t.Errorf("should have inline class")
	}
	if !strings.Contains(html, `id="my-id"`) {
		t.Errorf("should have inline id")
	}
	if !strings.Contains(html, `data-testid="section-1"`) {
		t.Errorf("should have data attribute")
	}
}

func TestIntegration_StaticAssetsCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md":              "# Page",
		"assets/style.css":     "body { color: red; }",
		"assets/images/bg.png": "fake png data",
	})

	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check static assets are copied
	cssContent, err := os.ReadFile(filepath.Join(buildDir, "assets/style.css"))
	if err != nil {
		t.Errorf("CSS file should be copied: %v", err)
	} else if string(cssContent) != "body { color: red; }" {
		t.Errorf("CSS content mismatch")
	}

	_, err = os.ReadFile(filepath.Join(buildDir, "assets/images/bg.png"))
	if err != nil {
		t.Errorf("PNG file should be copied: %v", err)
	}
}

func TestIntegration_TitleExtraction(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# My Page Title

Some content.
`,
	})

	gen := createGenerator(nil)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if !strings.Contains(string(content), "<title>My Page Title</title>") {
		t.Errorf("title tag should contain page title, got:\n%s", content)
	}
}

func TestIntegration_InvalidStyleConfigReturnsError(t *testing.T) {
	invalidConfig := `{
		"elements": {
			"invalid_key": "some-class"
		}
	}`

	_, err := styling.ParseConfig([]byte(invalidConfig))
	if err == nil {
		t.Error("ParseConfig should return error for invalid keys")
	}

	if !strings.Contains(err.Error(), "invalid_key") {
		t.Errorf("error should mention invalid key, got: %v", err)
	}

	if !strings.Contains(err.Error(), "heading1") {
		t.Errorf("error should list valid keys, got: %v", err)
	}
}

func TestIntegration_AllMarkdownElements(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Heading 1

## Heading 2

### Heading 3

Regular paragraph text.

[A link](https://example.com)

![An image](image.png)

> A blockquote

- Unordered list item 1
- Unordered list item 2

1. Ordered list item 1
2. Ordered list item 2

` + "```go\nfunc main() {}\n```" + `

Inline ` + "`code`" + ` here.

**Bold** and *italic* text.

---

| Table | Header |
|-------|--------|
| Cell  | Data   |
`,
		// Add the image file so validator doesn't fail
		"image.png": "fake png data",
	})

	styleConfig := &styling.Config{
		Elements: map[string]string{
			"heading1":   "h1-style",
			"heading2":   "h2-style",
			"heading3":   "h3-style",
			"paragraph":  "p-style",
			"link":       "link-style",
			"image":      "img-style",
			"blockquote": "quote-style",
			"list":       "list-style",
		},
		Contexts: make(map[string]map[string]string),
	}

	gen := createGenerator(styleConfig)
	err := gen.Generate(contentDir, buildDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)

	// Verify all styled elements
	checks := []struct {
		desc     string
		contains string
	}{
		{"H1 styled", `class="h1-style"`},
		{"H2 styled", `class="h2-style"`},
		{"H3 styled", `class="h3-style"`},
		{"Link styled", `class="link-style"`},
		{"Image styled", `class="img-style"`},
		{"Blockquote styled", `class="quote-style"`},
		{"List styled", `class="list-style"`},
		{"Code block present", `<pre><code`},
		{"Inline code", `<code>`},
		{"Bold text", `<strong>`},
		{"Italic text", `<em>`},
		{"Table present", `<table>`},
		{"Horizontal rule", `<hr`},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.contains) {
			t.Errorf("%s: expected %q in output", check.desc, check.contains)
		}
	}
}
