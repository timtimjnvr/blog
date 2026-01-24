package substitution

import (
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
)

func TestTitleSubstituter_Placeholder(t *testing.T) {
	sub := &TitleSubstituter{}
	if sub.Placeholder() != "{{title}}" {
		t.Errorf("Placeholder() = %q, want %q", sub.Placeholder(), "{{title}}")
	}
}

func TestTitleSubstituter_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "extracts H1 title",
			source:   "# Hello World\n\nContent here",
			expected: "Hello World",
		},
		{
			name:     "extracts title from middle of document",
			source:   "Some text\n# Title Here\n\nMore content",
			expected: "Title Here",
		},
		{
			name:     "returns Untitled when no H1",
			source:   "No heading here",
			expected: "Untitled",
		},
		{
			name:     "ignores H2 headings",
			source:   "## Only H2\n\nNo H1 present",
			expected: "Untitled",
		},
		{
			name:     "handles empty source",
			source:   "",
			expected: "Untitled",
		},
		{
			name:     "extracts first H1 when multiple exist",
			source:   "# First\n\n# Second",
			expected: "First",
		},
		{
			name:     "handles title with special characters",
			source:   "# Hello, World! (2024)",
			expected: "Hello, World! (2024)",
		},
	}

	sub := &TitleSubstituter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.PageContext{
				Source: []byte(tt.source),
			}
			result := sub.Resolve(ctx)
			if result != tt.expected {
				t.Errorf("Resolve() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple title", "# Hello World\n\nContent", "Hello World"},
		{"title with spaces", "#    Spaced Title   ", "Spaced Title   "},
		{"no title", "Just text", "Untitled"},
		{"H2 only", "## Not H1", "Untitled"},
		{"empty", "", "Untitled"},
		{"title in middle", "text\n# Middle Title\nmore", "Middle Title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTitle([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("ExtractTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
