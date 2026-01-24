package substitution

import (
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
)

func TestContentSubstituter_Placeholder(t *testing.T) {
	sub := &ContentSubstituter{}
	if sub.Placeholder() != "{{content}}" {
		t.Errorf("Placeholder() = %q, want %q", sub.Placeholder(), "{{content}}")
	}
}

func TestContentSubstituter_Resolve(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		relPath     string
		expected    string
	}{
		{
			name:        "converts md link to html",
			htmlContent: `<a href="posts/index.md">Articles</a>`,
			relPath:     "home.md",
			expected:    `<a href="posts/index.html">Articles</a>`,
		},
		{
			name:        "preserves content without md links",
			htmlContent: `<p>Hello World</p>`,
			relPath:     "home.md",
			expected:    `<p>Hello World</p>`,
		},
		{
			name:        "handles posts directory link conversion",
			htmlContent: `<a href="hello.md">Hello</a>`,
			relPath:     "posts/index.md",
			expected:    `<a href="../post/hello.html">Hello</a>`,
		},
	}

	sub := &ContentSubstituter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.PageContext{
				HTMLContent: tt.htmlContent,
				RelPath:     tt.relPath,
			}
			result := sub.Resolve(ctx)
			if result != tt.expected {
				t.Errorf("Resolve() = %q, want %q", result, tt.expected)
			}
		})
	}
}
