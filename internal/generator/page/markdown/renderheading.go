package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// HeadingRenderer renders headings with a clickable anchor link.
type HeadingRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *HeadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *HeadingRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		if id, ok := n.AttributeString("id"); ok {
			_, _ = w.WriteString(`<a href="#`)
			_, _ = w.Write(util.EscapeHTML(id.([]byte)))
			_, _ = w.WriteString(`" class="heading-anchor">#</a>`)
		}
	} else {
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}
