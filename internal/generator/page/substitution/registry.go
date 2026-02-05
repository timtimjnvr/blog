package substitution

import (
	"fmt"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page/substitution/content"
	"github.com/timtimjnvr/blog/internal/generator/page/substitution/title"
)

// Registry manages substitutions and applies them to templates
type Registry struct {
	substitutions []Substituter
}

// NewRegistry creates a new substitution registry
func NewRegistry() *Registry {
	return &Registry{
		substitutions: []Substituter{
			content.NewSubstituer(),
			title.NewSubstituer(),
		},
	}
}

// Apply applies all registered substitutions in the template at placeholder with content value resolved
func (r Registry) Apply(template, content string) (string, error) {
	result := template
	for _, s := range r.substitutions {
		resolution, err := s.Resolve(content)
		if err != nil {
			return "", fmt.Errorf("failed to resolve substitution: %v", err)
		}
		result = strings.ReplaceAll(result, s.Placeholder(), resolution)
	}
	return result, nil
}
