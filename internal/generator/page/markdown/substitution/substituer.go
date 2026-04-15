package substitution

type Substituer interface {
	Placeholder() string

	Resolve() (string, error)
}
