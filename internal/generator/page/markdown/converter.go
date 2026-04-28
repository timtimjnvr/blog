package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Converter wraps goldmark for markdown to HTML conversion
type Converter struct {
	md goldmark.Markdown
}

// NewConverter creates a new markdown converter with GFM extensions.
// If styleConfig is nil, no custom styling is applied.
// The context parameter allows context-specific styling (e.g., "post").
func NewConverter() (*Converter, error) {
	compiler, err := newD2Compiler()
	if err != nil {
		return nil, err
	}

	return &Converter{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAttribute(),
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				// Allow raw HTML injected during markdown substitutions (e.g. article listing).
				html.WithUnsafe(),
				// Each renderer handles a distinct node kind; priority only matters
				// relative to goldmark's built-in renderers (priority 1000).
				renderer.WithNodeRenderers(
					util.Prioritized(&HeadingRenderer{}, 100),
					util.Prioritized(&D2Renderer{compiler: compiler}, 100),
				),
			),
		),
	}, nil
}

// Convert converts markdown source to HTML.
func (c *Converter) Convert(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert(source, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
