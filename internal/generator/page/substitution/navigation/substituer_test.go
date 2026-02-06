package navigation

import (
	"strings"
	"testing"
)

func TestSubstituer_Placeholder(t *testing.T) {
	s := NewSubstituer(nil, "")
	if got := s.Placeholder(); got != "{{navigation}}" {
		t.Errorf("Placeholder() = %q, want %q", got, "{{navigation}}")
	}
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name           string
		sections       []string
		currentSection string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "from root with no sections",
			sections:       []string{},
			currentSection: "",
			wantContains:   []string{`href="index.html"`, "Accueil", "<nav"},
		},
		{
			name:           "from root with sections",
			sections:       []string{"posts", "about"},
			currentSection: "",
			wantContains: []string{
				`href="index.html"`,
				`href="posts/index.html"`,
				`href="about/index.html"`,
				"Accueil",
				"Posts",
				"About",
			},
		},
		{
			name:           "from section with sections",
			sections:       []string{"posts", "about"},
			currentSection: "posts",
			wantContains: []string{
				`href="../index.html"`,
				`href="../posts/index.html"`,
				`href="../about/index.html"`,
				"Accueil",
				"Posts",
				"About",
			},
			wantNotContain: []string{
				`href="index.html"`,
			},
		},
		{
			name:           "from about section",
			sections:       []string{"posts", "about"},
			currentSection: "about",
			wantContains: []string{
				`href="../index.html"`,
				`href="../posts/index.html"`,
				`href="../about/index.html"`,
			},
		},
		{
			name:           "from nested section",
			sections:       []string{"posts"},
			currentSection: "blog/2024",
			wantContains: []string{
				`href="../../index.html"`,
				`href="../../posts/index.html"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubstituer(tt.sections, tt.currentSection)
			got, err := s.Resolve("")
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			for _, substr := range tt.wantContains {
				if !strings.Contains(got, substr) {
					t.Errorf("Resolve() should contain %q, got:\n%s", substr, got)
				}
			}
			for _, substr := range tt.wantNotContain {
				if strings.Contains(got, substr) {
					t.Errorf("Resolve() should not contain %q, got:\n%s", substr, got)
				}
			}
		})
	}
}

func TestRelativePrefix(t *testing.T) {
	tests := []struct {
		section string
		want    string
	}{
		{"", ""},
		{"posts", "../"},
		{"about", "../"},
		{"blog/2024", "../../"},
		{"a/b/c", "../../../"},
	}
	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			if got := relativePrefix(tt.section); got != tt.want {
				t.Errorf("relativePrefix(%q) = %q, want %q", tt.section, got, tt.want)
			}
		})
	}
}

func TestSubstituer_Resolve_DisplayNameCapitalized(t *testing.T) {
	s := NewSubstituer([]string{"posts"}, "")
	got, err := s.Resolve("")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if !strings.Contains(got, "Posts") {
		t.Errorf("section name should be capitalized, got:\n%s", got)
	}
}
