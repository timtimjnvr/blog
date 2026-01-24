package substitution

import (
	"strings"

	"github.com/timtimjnvr/blog/internal/context"
)

// Substituter is implemented by all substitutions
type Substituter[T context.Context] interface {
	Placeholder() string
	Resolve(ctx T) string
}

// Registry manages substitutions and applies them to templates
type Registry[T context.Context] struct {
	substitutions []Substituter[T]
}

// NewRegistry creates a new substitution registry
func NewRegistry[T context.Context]() *Registry[T] {
	return &Registry[T]{}
}

// Register adds a substitution to the registry
func (r *Registry[T]) Register(s Substituter[T]) {
	r.substitutions = append(r.substitutions, s)
}

// Apply applies all registered substitutions to the template
func (r *Registry[T]) Apply(template string, ctx T) string {
	result := template
	for _, s := range r.substitutions {
		result = strings.ReplaceAll(result, s.Placeholder(), s.Resolve(ctx))
	}
	return result
}
