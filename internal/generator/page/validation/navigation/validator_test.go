package navigation

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator([]string{"posts", "about"})
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if len(v.sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(v.sections))
	}
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		sections   []string
		html       string
		wantErrors int
		wantMsg    []string
	}{
		{
			name:     "valid nav with all sections from root",
			sections: []string{"posts", "about"},
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="index.html">Accueil</a>
					<a href="posts/index.html">Posts</a>
					<a href="about/index.html">About</a>
				</nav>
				<p>Content</p>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name:     "valid nav with all sections from section depth",
			sections: []string{"posts", "about"},
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="../index.html">Accueil</a>
					<a href="../posts/index.html">Posts</a>
					<a href="../about/index.html">About</a>
				</nav>
				<p>Content</p>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name:       "missing nav element entirely",
			sections:   []string{"posts"},
			html:       `<html><body><p>No nav here</p></body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"missing <nav> element"},
		},
		{
			name:     "nav missing a section link",
			sections: []string{"posts", "about"},
			html: `<html><body>
				<nav>
					<a href="index.html">Accueil</a>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 2,
			wantMsg:    []string{"missing link to section \"about\"", "missing display name \"About\""},
		},
		{
			name:     "nav missing home link",
			sections: []string{"posts"},
			html: `<html><body>
				<nav>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 2,
			wantMsg:    []string{"missing home link (Accueil)", "missing home href"},
		},
		{
			name:     "nav with home but wrong display name",
			sections: []string{},
			html: `<html><body>
				<nav>
					<a href="index.html">Home</a>
				</nav>
			</body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"missing home link (Accueil)"},
		},
		{
			name:     "empty sections only requires home",
			sections: []string{},
			html: `<html><body>
				<nav>
					<a href="index.html">Accueil</a>
				</nav>
			</body></html>`,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(tt.sections)
			errs := v.Validate("test.html", "/build", []byte(tt.html))

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, want %d: %v", len(errs), tt.wantErrors, errs)
			}

			if len(tt.wantMsg) > 0 {
				allErrs := ""
				for _, e := range errs {
					allErrs += e.Error() + "\n"
				}
				for _, msg := range tt.wantMsg {
					if !strings.Contains(allErrs, msg) {
						t.Errorf("expected error containing %q, got:\n%s", msg, allErrs)
					}
				}
			}
		})
	}
}
