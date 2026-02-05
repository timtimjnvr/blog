package substitution

type Substituter interface {
	// The pattern to be replaced by the Resolve return
	Placeholder() string
	// Returns the new content with substitutions made
	Resolve(content string) (string, error)
}
