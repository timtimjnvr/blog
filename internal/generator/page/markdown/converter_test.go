package markdown

import (
	"strings"
	"testing"

	"github.com/timtimjnvr/blog/internal/generator/page/styling"
)

func TestNewConverter(t *testing.T) {
	converter := NewConverter(nil, "")
	if converter == nil {
		t.Fatal("NewConverter returned nil")
	}
	if converter.md == nil {
		t.Fatal("converter.md is nil")
	}
}

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "converts heading",
			input:    "# Hello World",
			contains: []string{"<h1>Hello World</h1>"},
		},
		{
			name:     "converts paragraph",
			input:    "This is a paragraph.",
			contains: []string{"<p>This is a paragraph.</p>"},
		},
		{
			name:     "converts link",
			input:    "[Link](https://example.com)",
			contains: []string{`<a href="https://example.com">Link</a>`},
		},
		{
			name:     "converts bold text",
			input:    "**bold**",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "converts italic text",
			input:    "*italic*",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "converts code block with GFM",
			input:    "```go\nfunc main() {}\n```",
			contains: []string{`<pre><code class="language-go">`},
		},
		{
			name:     "converts inline code",
			input:    "Use `fmt.Println`",
			contains: []string{"<code>fmt.Println</code>"},
		},
		{
			name:     "converts unordered list",
			input:    "- item 1\n- item 2",
			contains: []string{"<ul>", "<li>item 1</li>", "<li>item 2</li>"},
		},
		{
			name:     "converts ordered list",
			input:    "1. first\n2. second",
			contains: []string{"<ol>", "<li>first</li>", "<li>second</li>"},
		},
		{
			name:     "handles empty input",
			input:    "",
			contains: []string{},
		},
		{
			name:     "converts GFM strikethrough",
			input:    "~~deleted~~",
			contains: []string{"<del>deleted</del>"},
		},
		{
			name:     "converts GFM table",
			input:    "| A | B |\n|---|---|\n| 1 | 2 |",
			contains: []string{"<table>", "<th>A</th>", "<td>1</td>"},
		},
	}

	converter := NewConverter(nil, "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Convert() result should contain %q, got %q", substr, result)
				}
			}
		})
	}
}

func TestConverter_Convert_ReturnsValidHTML(t *testing.T) {
	converter := NewConverter(nil, "")

	input := "# Title\n\nParagraph with **bold** and *italic*.\n\n- List item"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Check that result is not empty
	if len(result) == 0 {
		t.Error("Convert() returned empty result")
	}

	// Check basic structure
	if !strings.Contains(result, "<h1>") {
		t.Error("missing h1 tag")
	}
	if !strings.Contains(result, "<p>") {
		t.Error("missing p tag")
	}
	if !strings.Contains(result, "<ul>") {
		t.Error("missing ul tag")
	}
}

func TestConverter_InlineAttributes(t *testing.T) {
	converter := NewConverter(nil, "")

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "heading with class",
			input:    "# Title {.custom-class}",
			contains: []string{`<h1 class="custom-class"`, "Title"},
		},
		{
			name:     "heading with id",
			input:    "## Section {#my-section}",
			contains: []string{`<h2 id="my-section"`, "Section"},
		},
		{
			name:     "heading with class and id",
			input:    "### Header {.styled #header-id}",
			contains: []string{`class="styled"`, `id="header-id"`, "Header"},
		},
		{
			name:     "heading with multiple classes",
			input:    "# Big Title {.text-4xl .font-bold .text-red-500}",
			contains: []string{`class="text-4xl font-bold text-red-500"`, "Big Title"},
		},
		{
			name:     "heading with custom attribute",
			input:    "# Title {data-testid=main-title}",
			contains: []string{`data-testid="main-title"`, "Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Convert() result should contain %q, got %q", substr, result)
				}
			}
		})
	}
}

func TestConverter_InlineAttributesOverrideConfig(t *testing.T) {
	// Test that inline attributes take precedence over styles.json config
	config := &styling.Config{
		Elements: map[string]string{
			"heading1": "config-class",
		},
		Contexts: make(map[string]map[string]string),
	}

	converter := NewConverter(config, "")

	// Inline attribute should override config
	input := "# Title {.inline-class}"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `class="inline-class"`) {
		t.Errorf("inline class should be present, got %q", result)
	}
}

func TestConverter_ConfigAppliedWithoutInlineAttributes(t *testing.T) {
	config := &styling.Config{
		Elements: map[string]string{
			"heading1": "from-config",
			"link":     "link-style",
		},
		Contexts: make(map[string]map[string]string),
	}

	converter := NewConverter(config, "")

	input := "# Title\n\n[Link](https://example.com)"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `class="from-config"`) {
		t.Errorf("heading should have config class, got %q", result)
	}
	if !strings.Contains(result, `class="link-style"`) {
		t.Errorf("link should have config class, got %q", result)
	}
}

func TestConverter_ContextSpecificStyling(t *testing.T) {
	config := &styling.Config{
		Elements: map[string]string{
			"heading1": "global-style",
		},
		Contexts: map[string]map[string]string{
			"post": {
				"heading1": "post-style",
			},
		},
	}

	t.Run("uses global style without context", func(t *testing.T) {
		converter := NewConverter(config, "")
		result, err := converter.Convert([]byte("# Title"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, `class="global-style"`) {
			t.Errorf("should use global style, got %q", result)
		}
	})

	t.Run("uses context style with post context", func(t *testing.T) {
		converter := NewConverter(config, "post")
		result, err := converter.Convert([]byte("# Title"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, `class="post-style"`) {
			t.Errorf("should use post context style, got %q", result)
		}
	})
}
