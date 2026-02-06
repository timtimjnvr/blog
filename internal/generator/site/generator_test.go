package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/generator/page/styling"
)

func setupTestContent(t *testing.T, files map[string]string) (contentDir, buildDir string) {
	t.Helper()
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

func createTestGenerator(contentDir, buildDir string) *Generator {
	return NewGenerator().
		WithContentDir(contentDir).
		WithBuildDir(buildDir).
		WithStylingConfigPath("")
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if g.contentDir != "content/markdown" {
		t.Errorf("contentDir = %q, want %q", g.contentDir, "content/markdown")
	}
	if g.buildDir != "target/build" {
		t.Errorf("buildDir = %q, want %q", g.buildDir, "target/build")
	}
	if g.assetsDir != "content/assets" {
		t.Errorf("assetsDir = %q, want %q", g.assetsDir, "content/assets")
	}
	if g.scriptsDir != "scripts" {
		t.Errorf("scriptsDir = %q, want %q", g.scriptsDir, "scripts")
	}
	if g.optionalStylingConfigPath != "styles/styles.json" {
		t.Errorf("optionalStylingConfigPath = %q, want %q", g.optionalStylingConfigPath, "styles/styles.json")
	}
	if g.pageGeneratorFactory == nil {
		t.Error("pageGeneratorFactory should not be nil")
	}
}

func TestWithBuilders(t *testing.T) {
	g := NewGenerator().
		WithContentDir("/content").
		WithBuildDir("/build").
		WithAssetsDir("/assets").
		WithScriptsDir("/scripts").
		WithStylingConfigPath("/styles.json")

	if g.contentDir != "/content" {
		t.Errorf("contentDir = %q, want %q", g.contentDir, "/content")
	}
	if g.buildDir != "/build" {
		t.Errorf("buildDir = %q, want %q", g.buildDir, "/build")
	}
	if g.assetsDir != "/assets" {
		t.Errorf("assetsDir = %q, want %q", g.assetsDir, "/assets")
	}
	if g.scriptsDir != "/scripts" {
		t.Errorf("scriptsDir = %q, want %q", g.scriptsDir, "/scripts")
	}
	if g.optionalStylingConfigPath != "/styles.json" {
		t.Errorf("optionalStylingConfigPath = %q, want %q", g.optionalStylingConfigPath, "/styles.json")
	}
}

func TestWithPageGeneratorFactory(t *testing.T) {
	called := false
	factory := func(markdownPath, buildDir, section string, stylingConfig *styling.Config) PageGenerator {
		called = true
		return &fakePageGenerator{}
	}

	g := NewGenerator().WithPageGeneratorFactory(factory)
	g.pageGeneratorFactory("test.md", "/build", "", nil)

	if !called {
		t.Error("custom factory should have been called")
	}
}

func TestExtractSection(t *testing.T) {
	tests := []struct {
		name       string
		contentDir string
		filePath   string
		want       string
		wantErr    bool
	}{
		{
			name:       "file at root of content dir",
			contentDir: "content/markdown",
			filePath:   "content/markdown/index.md",
			want:       "",
		},
		{
			name:       "file in section subdirectory",
			contentDir: "content/markdown",
			filePath:   "content/markdown/posts/hello.md",
			want:       "posts",
		},
		{
			name:       "file in nested section",
			contentDir: "content/markdown",
			filePath:   "content/markdown/blog/2024/post.md",
			want:       "blog/2024",
		},
		{
			name:       "file in about section",
			contentDir: "content/markdown",
			filePath:   "content/markdown/about/index.md",
			want:       "about",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSection(tt.contentDir, tt.filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("extractSection(%q, %q) = %q, want %q", tt.contentDir, tt.filePath, got, tt.want)
			}
		})
	}
}

type fakePageGenerator struct {
	generateErr error
	validateErr error
	generated   bool
	validated   bool
}

func (f *fakePageGenerator) Generate() error {
	f.generated = true
	return f.generateErr
}

func (f *fakePageGenerator) Validate() error {
	f.validated = true
	return f.validateErr
}

func TestGenerate_LoadStylingConfigError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	// Create an invalid styling config file
	styleConfigPath := filepath.Join(t.TempDir(), "bad-styles.json")
	os.WriteFile(styleConfigPath, []byte(`{"elements": {"bad_key": "class"}}`), 0644)

	gen := createTestGenerator(contentDir, buildDir).WithStylingConfigPath(styleConfigPath)
	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error for invalid styling config")
	}
	if !strings.Contains(err.Error(), "styling") {
		t.Errorf("error should mention styling, got %q", err.Error())
	}
}

func TestGenerate_NonExistentContentDirError(t *testing.T) {
	buildDir := t.TempDir()
	gen := createTestGenerator("/nonexistent/content", buildDir).
		WithAssetsDir("/nonexistent/assets").
		WithScriptsDir("/nonexistent/scripts")

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error for non-existent content dir")
	}
}

func TestGenerate_NonMdFilesReportError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":  "# Title",
		"extra.txt": "not markdown",
	})

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))

	// Create empty assets/scripts dirs so copyDir doesn't fail
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error for non-md file in content dir")
	}
	if !strings.Contains(err.Error(), "wrong extension") {
		t.Errorf("error should mention wrong extension, got %q", err.Error())
	}
}

func TestGenerate_PageGeneratorError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, buildDir, section string, stylingConfig *styling.Config) PageGenerator {
		return &fakePageGenerator{generateErr: fmt.Errorf("page generation failed")}
	}

	gen := createTestGenerator(contentDir, buildDir).
		WithPageGeneratorFactory(factory).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error when page generator fails")
	}
	if !strings.Contains(err.Error(), "page generation failed") {
		t.Errorf("error should contain page generator message, got %q", err.Error())
	}
}

func TestGenerate_ValidationError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, buildDir, section string, stylingConfig *styling.Config) PageGenerator {
		return &fakePageGenerator{validateErr: fmt.Errorf("validation failed")}
	}

	gen := createTestGenerator(contentDir, buildDir).
		WithPageGeneratorFactory(factory).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error when validation fails")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should contain validation message, got %q", err.Error())
	}
}

func TestValidate_NoPages(t *testing.T) {
	g := NewGenerator()
	err := g.Validate()
	if err != nil {
		t.Errorf("Validate() with no pages should return nil, got %v", err)
	}
}

func TestValidate_AllPass(t *testing.T) {
	g := NewGenerator()
	g.pagesGenerators = []PageGenerator{
		&fakePageGenerator{},
		&fakePageGenerator{},
	}
	err := g.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidate_WithErrors(t *testing.T) {
	g := NewGenerator()
	g.pagesGenerators = []PageGenerator{
		&fakePageGenerator{},
		&fakePageGenerator{validateErr: fmt.Errorf("broken link")},
		&fakePageGenerator{validateErr: fmt.Errorf("missing image")},
	}
	err := g.Validate()
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "broken link") {
		t.Errorf("error should contain 'broken link', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing image") {
		t.Errorf("error should contain 'missing image', got %q", err.Error())
	}
}

func TestIntegration_BasicSiteGeneration(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":        "# Welcome\n\nThis is the homepage.\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":  "# All Posts\n\n- [First Post](first.md)\n- [Second Post](second.md)\n",
		"posts/first.md":  "# First Post\n\nContent of the first post.\n\n[Back to posts](index.md)\n",
		"posts/second.md": "# Second Post\n\nContent of the second post.\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

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

	scriptsDir := t.TempDir()
	scriptContent := `(function() { console.log("test"); })();`
	os.WriteFile(filepath.Join(scriptsDir, "test.js"), []byte(scriptContent), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		WithScriptsDir(scriptsDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets"))
	os.MkdirAll(gen.assetsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

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

	scriptsDir := t.TempDir()
	darkModeScript := `(function() {
  const STORAGE_KEY = 'theme';
  function toggleTheme() { /* toggle */ }
  window.toggleTheme = toggleTheme;
})();`
	os.WriteFile(filepath.Join(scriptsDir, "dark-mode.js"), []byte(darkModeScript), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		WithScriptsDir(scriptsDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets"))
	os.MkdirAll(gen.assetsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

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
		"index.md":         "# Home\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":   "# Posts\n\n[Back to home](../index.md)\n[Read article](article.md)\n",
		"posts/article.md": "# Article\n\n[Back to index](index.md)\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	homeContent, _ := os.ReadFile(filepath.Join(buildDir, "index.html"))
	if !strings.Contains(string(homeContent), `href="posts/index.html"`) {
		t.Errorf("index should link to posts/index.html, got:\n%s", homeContent)
	}

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

	assetsDir := t.TempDir()
	os.MkdirAll(filepath.Join(assetsDir, "images"), 0755)
	os.WriteFile(filepath.Join(assetsDir, "style.css"), []byte("body { color: red; }"), 0644)
	os.WriteFile(filepath.Join(assetsDir, "images/bg.png"), []byte("fake png data"), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(assetsDir).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

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
		"page.md": "# My Page Title\n\nSome content.\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

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

func TestIntegration_StyleConfigApplied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"page.md": "# Main Title\n\nA paragraph with a [link](https://example.com).\n\n> A blockquote\n\n- List item\n",
	})

	styleConfigPath := filepath.Join(t.TempDir(), "styles.json")
	os.WriteFile(styleConfigPath, []byte(`{
		"elements": {
			"heading1": "custom-h1-class",
			"paragraph": "custom-p-class",
			"link": "custom-link-class",
			"blockquote": "custom-quote-class",
			"list": "custom-list-class"
		}
	}`), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		WithStylingConfigPath(styleConfigPath).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

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
		"page.md": "# Title {.my-class #my-id}\n\n## Section {data-testid=section-1}\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		WithAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		WithScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	os.MkdirAll(gen.assetsDir, 0755)
	os.MkdirAll(gen.scriptsDir, 0755)

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

func TestIntegration_InvalidStyleConfigReturnsError(t *testing.T) {
	invalidConfig := `{"elements": {"invalid_key": "some-class"}}`
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
