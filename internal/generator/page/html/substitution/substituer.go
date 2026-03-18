package substitution

type Substituer interface {
	// The pattern to be replaced by the Resolve return
	Placeholder() string

	// Returns the new content with substitutions made
	Resolve(content string) (string, error)
}
