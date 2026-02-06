package content

import (
	"strings"
	"testing"
)

func TestSubstituer_Placeholder(t *testing.T) {
	s := NewSubstituer()
	if got := s.Placeholder(); got != "{{content}}" {
		t.Errorf("Placeholder() = %q, want %q", got, "{{content}}")
	}
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		contains    []string
		notContains []string
	}{
		{
			name:        "converts md links to html links",
			htmlContent: `<a href="other.md">Other</a>`,
			contains:    []string{`href="other.html"`},
			notContains: []string{`other.md`},
		},
		{
			name:        "leaves external links unchanged",
			htmlContent: `<a href="https://example.com">External</a>`,
			contains:    []string{`href="https://example.com"`},
		},
		{
			name:        "leaves non-md links unchanged",
			htmlContent: `<a href="image.png">Image</a>`,
			contains:    []string{`href="image.png"`},
		},
		{
			name:        "handles empty content",
			htmlContent: "",
			contains:    []string{},
		},
		{
			name:        "handles content with no links",
			htmlContent: `<p>Just text</p>`,
			contains:    []string{"<p>Just text</p>"},
		},
	}

	s := NewSubstituer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Resolve(tt.htmlContent)
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("Resolve() result should contain %q, got %q", substr, got)
				}
			}
			for _, substr := range tt.notContains {
				if strings.Contains(got, substr) {
					t.Errorf("Resolve() result should not contain %q, got %q", substr, got)
				}
			}
		})
	}
}
