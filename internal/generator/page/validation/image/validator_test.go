package image

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", v.Timeout)
	}
	if v.SkipExternal {
		t.Error("SkipExternal should be false by default")
	}
}

func TestValidator_ValidateLocalImage(t *testing.T) {
	// Create temp directory structure
	buildDir := t.TempDir()
	imgDir := filepath.Join(buildDir, "assets", "images")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		t.Fatalf("failed to create image dir: %v", err)
	}

	// Create a test image file
	imgPath := filepath.Join(imgDir, "test.png")
	if err := os.WriteFile(imgPath, []byte("fake image"), 0644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	// Create HTML file path
	htmlDir := filepath.Join(buildDir, "post")
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
			name:      "valid relative image",
			html:      `<img src="../assets/images/test.png" alt="Test">`,
			wantError: false,
		},
		{
			name:      "missing relative image",
			html:      `<img src="../assets/images/missing.png" alt="Missing">`,
			wantError: true,
		},
		{
			name:      "valid absolute image",
			html:      `<img src="/assets/images/test.png" alt="Test">`,
			wantError: false,
		},
		{
			name:      "missing absolute image",
			html:      `<img src="/assets/images/missing.png" alt="Missing">`,
			wantError: true,
		},
		{
			name:      "no images",
			html:      `<p>No images here</p>`,
			wantError: false,
		},
		{
			name:      "multiple images mixed",
			html:      `<img src="../assets/images/test.png"><img src="../assets/images/missing.png">`,
			wantError: true,
		},
	}

	v := NewValidator()

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

func TestValidator_ValidateExternalImage(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid.png" {
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
			name:      "valid external image",
			html:      `<img src="` + server.URL + `/valid.png" alt="Valid">`,
			wantError: false,
		},
		{
			name:      "missing external image",
			html:      `<img src="` + server.URL + `/missing.png" alt="Missing">`,
			wantError: true,
		},
	}

	v := NewValidator()
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

func TestValidator_SkipExternal(t *testing.T) {
	buildDir := t.TempDir()
	htmlPath := filepath.Join(buildDir, "test.html")

	// This URL would fail if actually checked
	html := `<img src="http://invalid.invalid/image.png" alt="Invalid">`

	v := NewValidator()
	v.SkipExternal = true

	errs := v.Validate(htmlPath, buildDir, []byte(html))
	if len(errs) > 0 {
		t.Errorf("expected no errors when SkipExternal=true, got %v", errs)
	}
}

func TestIsExternalURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://example.com/image.png", true},
		{"https://example.com/image.png", true},
		{"../assets/image.png", false},
		{"/assets/image.png", false},
		{"image.png", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isExternalURL(tt.url)
			if result != tt.expected {
				t.Errorf("isExternalURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}
