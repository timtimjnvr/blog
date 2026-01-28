package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewScriptValidator(t *testing.T) {
	v := NewScriptValidator()
	if v == nil {
		t.Fatal("NewScriptValidator returned nil")
	}
}

func TestScriptValidator_ValidateLocalScript(t *testing.T) {
	buildDir := t.TempDir()
	scriptsDir := filepath.Join(buildDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	// Create a valid JS file
	validJS := `(function() { console.log("hello"); })();`
	if err := os.WriteFile(filepath.Join(scriptsDir, "valid.js"), []byte(validJS), 0644); err != nil {
		t.Fatalf("failed to create valid.js: %v", err)
	}

	htmlPath := filepath.Join(buildDir, "test.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid script exists",
			html:      `<script src="/scripts/valid.js"></script>`,
			wantError: false,
		},
		{
			name:      "missing script",
			html:      `<script src="/scripts/missing.js"></script>`,
			wantError: true,
		},
		{
			name:      "no scripts",
			html:      `<p>No scripts here</p>`,
			wantError: false,
		},
		{
			name:      "external script ignored",
			html:      `<script src="https://cdn.example.com/lib.js"></script>`,
			wantError: false,
		},
	}

	v := NewScriptValidator()

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

func TestScriptValidator_ValidateJSSyntax(t *testing.T) {
	buildDir := t.TempDir()
	scriptsDir := filepath.Join(buildDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	tests := []struct {
		name      string
		jsContent string
		wantError bool
	}{
		{
			name:      "valid IIFE",
			jsContent: `(function() { var x = 1; })();`,
			wantError: false,
		},
		{
			name:      "valid arrow function",
			jsContent: `const fn = () => { return 42; };`,
			wantError: false,
		},
		{
			name:      "valid ES6 class",
			jsContent: `class Foo { constructor() { this.x = 1; } }`,
			wantError: false,
		},
		{
			name:      "syntax error - missing bracket",
			jsContent: `function foo( { return 1; }`,
			wantError: true,
		},
		{
			name:      "syntax error - invalid token",
			jsContent: `var x = @invalid;`,
			wantError: true,
		},
		{
			name:      "syntax error - unclosed string",
			jsContent: `var x = "unclosed;`,
			wantError: true,
		},
	}

	v := NewScriptValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(scriptsDir, "test.js")
			if err := os.WriteFile(scriptPath, []byte(tt.jsContent), 0644); err != nil {
				t.Fatalf("failed to write test script: %v", err)
			}

			html := `<script src="/scripts/test.js"></script>`
			htmlPath := filepath.Join(buildDir, "test.html")

			errs := v.Validate(htmlPath, buildDir, []byte(html))
			hasError := len(errs) > 0
			if hasError != tt.wantError {
				t.Errorf("Validate() hasError = %v, want %v, errors: %v", hasError, tt.wantError, errs)
			}
		})
	}
}

func TestScriptValidator_RelativePath(t *testing.T) {
	buildDir := t.TempDir()

	// Create nested structure: build/post/article.html and build/scripts/app.js
	postDir := filepath.Join(buildDir, "post")
	scriptsDir := filepath.Join(buildDir, "scripts")
	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post dir: %v", err)
	}
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	validJS := `console.log("test");`
	if err := os.WriteFile(filepath.Join(scriptsDir, "app.js"), []byte(validJS), 0644); err != nil {
		t.Fatalf("failed to create app.js: %v", err)
	}

	htmlPath := filepath.Join(postDir, "article.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "relative path from nested dir",
			html:      `<script src="../scripts/app.js"></script>`,
			wantError: false,
		},
		{
			name:      "absolute path from nested dir",
			html:      `<script src="/scripts/app.js"></script>`,
			wantError: false,
		},
	}

	v := NewScriptValidator()

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
