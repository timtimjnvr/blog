package markdown

import (
	"bytes"

	"github.com/timtimjnvr/blog/internal/styling"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// Converter wraps goldmark for markdown to HTML conversion
type Converter struct {
	md goldmark.Markdown
}

// NewConverter creates a new markdown converter with GFM extensions.
// If styleConfig is nil, no custom styling is applied.
// The context parameter allows context-specific styling (e.g., "post").
func NewConverter(styleConfig *styling.Config, context string) *Converter {
	parserOpts := []parser.Option{
		// Enable inline attribute syntax: {.class #id key=value}
		parser.WithAttribute(),
	}

	// Add style transformer if config is provided
	if styleConfig != nil {
		transformer := styling.NewTransformer(styleConfig, context)
		parserOpts = append(parserOpts,
			parser.WithASTTransformers(
				util.Prioritized(transformer, 100),
			),
		)
	}

	return &Converter{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(parserOpts...),
		),
	}
}

// Convert converts markdown source to HTML
func (c *Converter) Convert(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert(source, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
