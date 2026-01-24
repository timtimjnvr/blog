package substitution

import (
	"regexp"

	"github.com/timtimjnvr/blog/internal/context"
)

// TitleSubstituter resolves {{title}} placeholder
type TitleSubstituter struct{}

func (t *TitleSubstituter) Placeholder() string {
	return "{{title}}"
}

func (t *TitleSubstituter) Resolve(ctx *context.PageContext) string {
	return ExtractTitle(ctx.GetSource())
}

// ExtractTitle extracts the first H1 heading from markdown content
func ExtractTitle(source []byte) string {
	re := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	match := re.FindSubmatch(source)
	if len(match) >= 2 {
		return string(match[1])
	}
	return "Untitled"
}
