package title

import (
	"testing"
)

func TestSubstituer_Placeholder(t *testing.T) {
	s := NewSubstituer()
	if got := s.Placeholder(); got != "{{title}}" {
		t.Errorf("Placeholder() = %q, want %q", got, "{{title}}")
	}
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "extracts title from h1 tag",
			content: `<h1>My Page Title</h1><p>Some content</p>`,
			want:    "My Page Title",
		},
		{
			name:    "extracts title from h1 with classes",
			content: `<h1 class="text-4xl font-bold">Styled Title</h1>`,
			want:    "Styled Title",
		},
		{
			name:    "extracts title from h1 with id",
			content: `<h1 id="main-title">ID Title</h1>`,
			want:    "ID Title",
		},
		{
			name:    "extracts first h1 when multiple exist",
			content: `<h1>First</h1><h1>Second</h1>`,
			want:    "First",
		},
		{
			name:    "returns error when no h1 found",
			content: `<h2>Not a title</h2><p>Content</p>`,
			wantErr: true,
		},
		{
			name:    "returns error for empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "returns error for content without headings",
			content: `<p>Just a paragraph</p>`,
			wantErr: true,
		},
	}

	s := NewSubstituer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Resolve(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Error("Resolve() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
