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

// NewRegistry creates a new substitution registry with default substituters
func NewRegistry() *Registry {
	return NewRegistryWithSubstituters(
		content.NewSubstituer(),
		title.NewSubstituer(),
	)
}

// NewRegistryWithSubstituters creates a registry with custom substituters
func NewRegistryWithSubstituters(subs ...Substituter) *Registry {
	return &Registry{
		substitutions: subs,
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
