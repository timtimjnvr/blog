package validation

import (
	"errors"

	"github.com/timtimjnvr/blog/internal/generator/page/html/validation/image"
	"github.com/timtimjnvr/blog/internal/generator/page/html/validation/link"
	"github.com/timtimjnvr/blog/internal/generator/page/html/validation/navigation"
	"github.com/timtimjnvr/blog/internal/generator/section"
)

// Registry manages validators and runs them on HTML content
type Registry struct {
	validators []Validator
}

// NewRegistry creates a validation registry with the navigation validator configured for the given sections
func NewRegistry(sections []section.Section) *Registry {
	return &Registry{
		validators: []Validator{
			link.NewValidator(),
			image.NewValidator(),
			navigation.NewValidator(sections),
		},
	}
}

// NewRegistryWithValidators creates a registry with custom validators
func NewRegistryWithValidators(validators ...Validator) *Registry {
	return &Registry{
		validators: validators,
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
