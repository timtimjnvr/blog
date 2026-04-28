package markdown

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

// d2Compiler compiles D2 source to SVG. It is intended to be created once
// and shared across all diagrams in a generation run, since NewRuler loads
// font data and is expensive.
type d2Compiler struct {
	ruler *textmeasure.Ruler
	ctx   context.Context
}

func newD2Compiler() (*d2Compiler, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("d2 ruler: %w", err)
	}
	ctx := log.With(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return &d2Compiler{ruler: ruler, ctx: ctx}, nil
}

func (c *d2Compiler) compile(code string, scale float64) (string, error) {
	diagram, _, err := d2lib.Compile(c.ctx, code, &d2lib.CompileOptions{
		Ruler: c.ruler,
		LayoutResolver: func(_ string) (d2graph.LayoutGraph, error) {
			return func(ctx context.Context, g *d2graph.Graph) error {
				return d2dagrelayout.Layout(ctx, g, nil)
			}, nil
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("d2 compile: %w", err)
	}

	renderOpts := &d2svg.RenderOpts{}
	if scale > 0 {
		renderOpts.Scale = &scale
	}
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return "", fmt.Errorf("d2 render: %w", err)
	}

	return string(svg), nil
}
