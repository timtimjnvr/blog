package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/site"
)

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

// Helper to create a generator with test directories
func createTestGenerator(contentDir, buildDir string) *site.Generator {
	return site.NewGenerator().
		WithContentDir(contentDir).
		WithBuildDir(buildDir).
		WithStylingConfigPath("") // No styling config by default
}

// Integration tests for the full site generation pipeline

func TestIntegration_BasicSiteGeneration(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": `# Welcome

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

	gen := createTestGenerator(contentDir, buildDir)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify files are generated
	expectedFiles := []string{
		"index.html",
		"posts/index.html",
		"posts/first.html",
		"posts/second.html",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(buildDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestIntegration_ScriptsCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": "# Page",
	})

	// Create scripts directory
	scriptsDir := t.TempDir()
	scriptContent := `(function() { console.log("test"); })();`
	if err := os.WriteFile(filepath.Join(scriptsDir, "test.js"), []byte(scriptContent), 0644); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	gen := createTestGenerator(contentDir, buildDir).WithScriptsDir(scriptsDir)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check script was copied
	copiedScript, err := os.ReadFile(filepath.Join(buildDir, "scripts/test.js"))
	if err != nil {
		t.Fatalf("script should be copied: %v", err)
	}

	if string(copiedScript) != scriptContent {
		t.Errorf("script content mismatch, got: %s", copiedScript)
	}
}

func TestIntegration_DarkModeScriptCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": "# Page",
	})

	// Create scripts directory with dark-mode.js
	scriptsDir := t.TempDir()
	darkModeScript := `(function() {
  const STORAGE_KEY = 'theme';
  function toggleTheme() { /* toggle */ }
  window.toggleTheme = toggleTheme;
})();`
	if err := os.WriteFile(filepath.Join(scriptsDir, "dark-mode.js"), []byte(darkModeScript), 0644); err != nil {
		t.Fatalf("failed to create dark-mode.js: %v", err)
	}

	gen := createTestGenerator(contentDir, buildDir).WithScriptsDir(scriptsDir)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check dark-mode.js was copied
	copiedScript, err := os.ReadFile(filepath.Join(buildDir, "scripts/dark-mode.js"))
	if err != nil {
		t.Fatalf("dark-mode.js should be copied: %v", err)
	}

	if !strings.Contains(string(copiedScript), "toggleTheme") {
		t.Errorf("dark-mode.js should contain toggleTheme function")
	}
}

func TestIntegration_LinkConversion(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": `# Home

[Go to posts](posts/index.md)
`,
		"posts/index.md": `# Posts

[Back to home](../index.md)
[Read article](article.md)
`,
		"posts/article.md": `# Article

[Back to index](index.md)
`,
	})

	gen := createTestGenerator(contentDir, buildDir)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check index.md links
	homeContent, _ := os.ReadFile(filepath.Join(buildDir, "index.html"))
	if !strings.Contains(string(homeContent), `href="posts/index.html"`) {
		t.Errorf("index should link to posts/index.html, got:\n%s", homeContent)
	}

	// Check posts/index.md links - relative paths are preserved with .md -> .html conversion
	postsContent, _ := os.ReadFile(filepath.Join(buildDir, "posts/index.html"))
	if !strings.Contains(string(postsContent), `href="../index.html"`) && !strings.Contains(string(postsContent), `href="index.html"`) {
		t.Errorf("posts/index should link to index.html, got:\n%s", postsContent)
	}
	if !strings.Contains(string(postsContent), `href="article.html"`) {
		t.Errorf("posts/index should link to article.html, got:\n%s", postsContent)
	}
}

func TestIntegration_StaticAssetsCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": "# Page",
	})

	// Create assets directory
	assetsDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(assetsDir, "images"), 0755); err != nil {
		t.Fatalf("failed to create images dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "style.css"), []byte("body { color: red; }"), 0644); err != nil {
		t.Fatalf("failed to create style.css: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "images/bg.png"), []byte("fake png data"), 0644); err != nil {
		t.Fatalf("failed to create bg.png: %v", err)
	}

	gen := createTestGenerator(contentDir, buildDir).WithAssetsDir(assetsDir)
	err := gen.Generate()
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

	gen := createTestGenerator(contentDir, buildDir)
	err := gen.Generate()
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

func TestIntegration_StyleConfigApplied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Main Title

A paragraph with a [link](https://example.com).

> A blockquote

- List item
`,
	})

	// Create style config file
	styleConfigPath := filepath.Join(t.TempDir(), "styles.json")
	styleConfig := `{
		"elements": {
			"heading1": "custom-h1-class",
			"paragraph": "custom-p-class",
			"link": "custom-link-class",
			"blockquote": "custom-quote-class",
			"list": "custom-list-class"
		}
	}`
	if err := os.WriteFile(styleConfigPath, []byte(styleConfig), 0644); err != nil {
		t.Fatalf("failed to write style config: %v", err)
	}

	gen := createTestGenerator(contentDir, buildDir).WithStylingConfigPath(styleConfigPath)
	err := gen.Generate()
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

func TestIntegration_InlineAttributesWithoutConfig(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": `# Title {.my-class #my-id}

## Section {data-testid=section-1}
`,
	})

	gen := createTestGenerator(contentDir, buildDir)
	err := gen.Generate()
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
