package section

// Section represents a top-level site section.
type Section struct {
	DirName     string // directory name (used for URL path construction)
	DisplayName string // display name shown in navigation (from # title in index.md)
}
