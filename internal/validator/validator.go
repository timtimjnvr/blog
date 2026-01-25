package validator

// ValidationError represents a validation failure
type ValidationError struct {
	File    string
	Message string
}

func (e ValidationError) Error() string {
	return e.File + ": " + e.Message
}

// Validator validates generated HTML content
type Validator interface {
	// Validate checks the HTML content and returns any validation errors
	// htmlPath is the path to the generated HTML file
	// buildDir is the root build directory for resolving relative paths
	Validate(htmlPath, buildDir string, content []byte) []ValidationError
}

// ValidationResult contains all validation errors
type ValidationResult struct {
	Errors []ValidationError
}

func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) Add(errs ...ValidationError) {
	r.Errors = append(r.Errors, errs...)
}

// Runner runs multiple validators
type Runner struct {
	validators []Validator
}

// NewRunner creates a new validation runner
func NewRunner() *Runner {
	return &Runner{}
}

// Register adds a validator to the runner
func (r *Runner) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs all validators on the given HTML file
func (r *Runner) Validate(htmlPath, buildDir string, content []byte) *ValidationResult {
	result := &ValidationResult{}
	for _, v := range r.validators {
		errs := v.Validate(htmlPath, buildDir, content)
		result.Add(errs...)
	}
	return result
}
