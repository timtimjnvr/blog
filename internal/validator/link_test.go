package validator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLinkValidator(t *testing.T) {
	v := NewLinkValidator()
	if v == nil {
		t.Fatal("NewLinkValidator returned nil")
	}
	if v.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", v.Timeout)
	}
	if v.SkipExternal {
		t.Error("SkipExternal should be false by default")
	}
}

func TestLinkValidator_ValidateLocalLink(t *testing.T) {
	// Create temp directory structure
	buildDir := t.TempDir()
	pagesDir := filepath.Join(buildDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		t.Fatalf("failed to create pages dir: %v", err)
	}

	// Create a test HTML file
	aboutPath := filepath.Join(pagesDir, "about.html")
	if err := os.WriteFile(aboutPath, []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("failed to create about page: %v", err)
	}

	// Create a directory with index.html
	postsDir := filepath.Join(buildDir, "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatalf("failed to create posts dir: %v", err)
	}
	indexPath := filepath.Join(postsDir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("failed to create index page: %v", err)
	}

	// Create HTML file path for testing
	htmlDir := filepath.Join(buildDir, "blog")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		t.Fatalf("failed to create html dir: %v", err)
	}
	htmlPath := filepath.Join(htmlDir, "test.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid relative link",
			html:      `<a href="../pages/about.html">About</a>`,
			wantError: false,
		},
		{
			name:      "missing relative link",
			html:      `<a href="../pages/missing.html">Missing</a>`,
			wantError: true,
		},
		{
			name:      "valid absolute link",
			html:      `<a href="/pages/about.html">About</a>`,
			wantError: false,
		},
		{
			name:      "missing absolute link",
			html:      `<a href="/pages/missing.html">Missing</a>`,
			wantError: true,
		},
		{
			name:      "valid directory link with index",
			html:      `<a href="/posts">Posts</a>`,
			wantError: false,
		},
		{
			name:      "valid directory link with trailing slash",
			html:      `<a href="/posts/">Posts</a>`,
			wantError: false,
		},
		{
			name:      "fragment only link",
			html:      `<a href="#section">Section</a>`,
			wantError: false,
		},
		{
			name:      "link with fragment",
			html:      `<a href="/pages/about.html#section">About Section</a>`,
			wantError: false,
		},
		{
			name:      "mailto link",
			html:      `<a href="mailto:test@example.com">Email</a>`,
			wantError: false,
		},
		{
			name:      "tel link",
			html:      `<a href="tel:+1234567890">Call</a>`,
			wantError: false,
		},
		{
			name:      "no links",
			html:      `<p>No links here</p>`,
			wantError: false,
		},
		{
			name:      "multiple links mixed",
			html:      `<a href="/pages/about.html">Valid</a><a href="/missing.html">Missing</a>`,
			wantError: true,
		},
	}

	v := NewLinkValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(htmlPath, buildDir, []byte(tt.html))
			hasError := len(errs) > 0
			if hasError != tt.wantError {
				t.Errorf("Validate() hasError = %v, want %v, errors: %v", hasError, tt.wantError, errs)
			}
		})
	}
}

func TestLinkValidator_ValidateExternalLink(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	buildDir := t.TempDir()
	htmlPath := filepath.Join(buildDir, "test.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid external link",
			html:      `<a href="` + server.URL + `/valid">Valid</a>`,
			wantError: false,
		},
		{
			name:      "missing external link",
			html:      `<a href="` + server.URL + `/missing">Missing</a>`,
			wantError: true,
		},
	}

	v := NewLinkValidator()
	v.Timeout = 5 * time.Second

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(htmlPath, buildDir, []byte(tt.html))
			hasError := len(errs) > 0
			if hasError != tt.wantError {
				t.Errorf("Validate() hasError = %v, want %v, errors: %v", hasError, tt.wantError, errs)
			}
		})
	}
}

func TestLinkValidator_SkipExternal(t *testing.T) {
	buildDir := t.TempDir()
	htmlPath := filepath.Join(buildDir, "test.html")

	// This URL would fail if actually checked
	html := `<a href="http://invalid.invalid/page">Invalid</a>`

	v := NewLinkValidator()
	v.SkipExternal = true

	errs := v.Validate(htmlPath, buildDir, []byte(html))
	if len(errs) > 0 {
		t.Errorf("expected no errors when SkipExternal=true, got %v", errs)
	}
}

func TestLinkValidator_SkipsJavascriptLinks(t *testing.T) {
	buildDir := t.TempDir()
	htmlPath := filepath.Join(buildDir, "test.html")

	html := `<a href="javascript:void(0)">Click</a>`

	v := NewLinkValidator()

	errs := v.Validate(htmlPath, buildDir, []byte(html))
	if len(errs) > 0 {
		t.Errorf("expected no errors for javascript link, got %v", errs)
	}
}
