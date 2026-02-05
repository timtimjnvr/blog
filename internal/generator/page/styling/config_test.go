package styling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig returned nil")
	}
	if config.Elements == nil {
		t.Error("Elements map is nil")
	}
	if config.Contexts == nil {
		t.Error("Contexts map is nil")
	}
}

func TestConfig_GetClasses(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		elementType string
		context     string
		want        string
	}{
		{
			name: "returns global element class",
			config: &Config{
				Elements: map[string]string{
					"heading1": "text-4xl font-bold",
				},
				Contexts: make(map[string]map[string]string),
			},
			elementType: "heading1",
			context:     "",
			want:        "text-4xl font-bold",
		},
		{
			name: "returns empty string for unknown element",
			config: &Config{
				Elements: map[string]string{},
				Contexts: make(map[string]map[string]string),
			},
			elementType: "unknown",
			context:     "",
			want:        "",
		},
		{
			name: "context override takes precedence",
			config: &Config{
				Elements: map[string]string{
					"heading1": "global-class",
				},
				Contexts: map[string]map[string]string{
					"post": {
						"heading1": "post-specific-class",
					},
				},
			},
			elementType: "heading1",
			context:     "post",
			want:        "post-specific-class",
		},
		{
			name: "falls back to global when context has no override",
			config: &Config{
				Elements: map[string]string{
					"heading1": "global-class",
				},
				Contexts: map[string]map[string]string{
					"post": {
						"paragraph": "post-paragraph",
					},
				},
			},
			elementType: "heading1",
			context:     "post",
			want:        "global-class",
		},
		{
			name: "returns empty when context does not exist",
			config: &Config{
				Elements: map[string]string{},
				Contexts: make(map[string]map[string]string),
			},
			elementType: "heading1",
			context:     "nonexistent",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetClasses(tt.elementType, tt.context)
			if got != tt.want {
				t.Errorf("GetClasses() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "parses valid config",
			json: `{
				"elements": {
					"heading1": "text-4xl font-bold",
					"paragraph": "text-base"
				},
				"contexts": {
					"post": {
						"heading1": "text-blue-900"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "rejects invalid element key",
			json: `{
				"elements": {
					"invalidkey": "some-class"
				}
			}`,
			wantErr: true,
		},
		{
			name: "rejects invalid context element key",
			json: `{
				"elements": {
					"heading1": "valid"
				},
				"contexts": {
					"post": {
						"notvalid": "some-class"
					}
				}
			}`,
			wantErr: true,
		},
		{
			name:    "accepts empty config",
			json:    ``,
			wantErr: false,
		},
		{
			name:    "accepts empty object",
			json:    `{}`,
			wantErr: false,
		},
		{
			name: "accepts all valid element types",
			json: `{
				"elements": {
					"heading1": "h1",
					"heading2": "h2",
					"heading3": "h3",
					"heading4": "h4",
					"heading5": "h5",
					"heading6": "h6",
					"paragraph": "p",
					"link": "a",
					"image": "img",
					"codeblock": "pre",
					"code": "code",
					"blockquote": "bq",
					"list": "ul",
					"listitem": "li"
				}
			}`,
			wantErr: false,
		},
		{
			name:    "rejects malformed json",
			json:    `{"elements": {invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseConfig([]byte(tt.json))

			if tt.wantErr {
				if err == nil {
					t.Error("ParseConfig() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ParseConfig() unexpected error: %v", err)
				}
				if config == nil {
					t.Error("ParseConfig() returned nil config")
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config passes",
			config: &Config{
				Elements: map[string]string{
					"heading1": "class",
					"link":     "class",
				},
				Contexts: map[string]map[string]string{
					"post": {"image": "class"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid element key fails",
			config: &Config{
				Elements: map[string]string{
					"h1": "class", // should be "heading1"
				},
				Contexts: make(map[string]map[string]string),
			},
			wantErr: true,
			errMsg:  "elements.h1",
		},
		{
			name: "invalid context key fails",
			config: &Config{
				Elements: make(map[string]string),
				Contexts: map[string]map[string]string{
					"post": {"div": "class"}, // "div" is not valid
				},
			},
			wantErr: true,
			errMsg:  "contexts.post.div",
		},
		{
			name: "error message lists valid keys",
			config: &Config{
				Elements: map[string]string{"bad": "class"},
				Contexts: make(map[string]map[string]string),
			},
			wantErr: true,
			errMsg:  "heading1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error should contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "styles.json")

		content := `{
			"elements": {
				"heading1": "text-4xl",
				"link": "text-blue-600"
			}
		}`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		config, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig() error: %v", err)
		}

		if config.Elements["heading1"] != "text-4xl" {
			t.Errorf("expected heading1 class 'text-4xl', got %q", config.Elements["heading1"])
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/styles.json")
		if err == nil {
			t.Error("LoadConfig() expected error for nonexistent file")
		}
	})

	t.Run("returns error for invalid config", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "styles.json")

		content := `{
			"elements": {
				"badkey": "class"
			}
		}`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadConfig(path)
		if err == nil {
			t.Error("LoadConfig() expected error for invalid config")
		}
	})
}
