package markdown

import (
	"strings"
	"testing"
)

func TestNewConverter(t *testing.T) {
	converter := NewConverter()
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

	converter := NewConverter()

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
	converter := NewConverter()

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
