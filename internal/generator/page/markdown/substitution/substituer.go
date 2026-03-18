package substitution

type Substituer interface {
	PlaceHolder() string

	Resolve() (string, error)
}
