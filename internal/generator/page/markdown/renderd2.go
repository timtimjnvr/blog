package markdown

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var d2ScaleAttr = regexp.MustCompile(`\bscale=(\d+(?:\.\w+)?)\b`)

type d2BlockAttrs struct {
	scale float64
}

// parseD2Attrs reads scale=N from the fenced block info string.
func parseD2Attrs(source []byte, n *ast.FencedCodeBlock) (d2BlockAttrs, error) {
	if n.Info == nil {
		return d2BlockAttrs{}, nil
	}
	info := n.Info.Segment.Value(source)
	var attrs d2BlockAttrs
	var err error
	if m := d2ScaleAttr.FindSubmatch(info); len(m) >= 2 {
		attrs.scale, err = strconv.ParseFloat(string(m[1]), 64)
		if err != nil {
			return attrs, err
		}
	}
	return attrs, nil
}

// D2Renderer renders d2 fenced code blocks as inline SVG.
// All other fenced code blocks are rendered as standard <pre><code> blocks.
type D2Renderer struct {
	compiler *d2Compiler
}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *D2Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *D2Renderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)

	if string(n.Language(source)) != "d2" {
		return r.renderCodeBlock(w, source, n, entering)
	}

	if entering {
		attrs, err := parseD2Attrs(source, n)
		if err != nil {
			return ast.WalkStop, err
		}
		code := extractCodeLines(source, n)
		svg, err := r.compiler.compile(code, attrs.scale)
		if err != nil {
			return ast.WalkStop, err
		}
		_, _ = w.WriteString(`<div style="width: 100%; display: flex; justify-content: center">`)
		_, _ = w.WriteString(svg)
		_, _ = w.WriteString("</div>\n")
	}
	return ast.WalkSkipChildren, nil
}

func (r *D2Renderer) renderCodeBlock(
	w util.BufWriter, source []byte, n *ast.FencedCodeBlock, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<pre><code")
		if lang := n.Language(source); lang != nil {
			_, _ = w.WriteString(` class="language-`)
			_, _ = w.Write(util.EscapeHTML(lang))
			_, _ = w.WriteString(`"`)
		}
		_ = w.WriteByte('>')
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			_, _ = w.Write(util.EscapeHTML(seg.Value(source)))
		}
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func extractCodeLines(source []byte, n *ast.FencedCodeBlock) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(source))
	}
	return b.String()
}
