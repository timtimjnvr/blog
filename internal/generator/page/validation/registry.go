package validation

import (
	"errors"

	"github.com/timtimjnvr/blog/internal/generator/page/validation/image"
	"github.com/timtimjnvr/blog/internal/generator/page/validation/link"
	"github.com/timtimjnvr/blog/internal/generator/page/validation/script"
)

// Registry manages validators and runs them on HTML content
type Registry struct {
	validators []Validator
}

// NewRegistry creates an empty validation registry
func NewRegistry() *Registry {
	return &Registry{
		validators: []Validator{},
	}
}

// NewDefaultRegistry creates a validation registry with default validators (image, script, link)
func NewDefaultRegistry() *Registry {
	return &Registry{
		validators: []Validator{
			image.NewValidator(),
			script.NewValidator(),
			link.NewValidator(),
		},
	}
}

// Register adds a validator to the registry
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs all registered validators on the given HTML content
func (r *Registry) Validate(htmlPath, buildDir string, content []byte) error {
	var errs []error
	for _, v := range r.validators {
		errs = append(errs, v.Validate(htmlPath, buildDir, content)...)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
