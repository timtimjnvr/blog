package substitution

import (
	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/markdown"
)

// ContentSubstituter resolves {{content}} placeholder
type ContentSubstituter struct{}

func (c *ContentSubstituter) Placeholder() string {
	return "{{content}}"
}

func (c *ContentSubstituter) Resolve(ctx *context.PageContext) string {
	html := markdown.ConvertMdLinksToHtml(ctx.GetHTMLContent(), ctx.GetRelPath())
	html = markdown.ConvertAssetPaths(html, ctx.GetRelPath())
	return html
}
