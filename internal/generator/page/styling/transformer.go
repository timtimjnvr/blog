package styling

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Transformer is a Goldmark ASTTransformer that adds CSS classes to AST nodes
// based on a Config.
type Transformer struct {
	config  *Config
	context string
}

// NewTransformer creates a new Transformer with the given config.
// The context parameter is optional and allows context-specific styling (e.g., "post").
func NewTransformer(config *Config, context string) *Transformer {
	if config == nil {
		config = &Config{
			Elements: make(map[string]string),
			Contexts: make(map[string]map[string]string),
		}
	}
	return &Transformer{
		config:  config,
		context: context,
	}
}

// Transform implements parser.ASTTransformer.
func (t *Transformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// func does not return error
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		var elementType string

		switch node := n.(type) {
		case *ast.Heading:
			elementType = t.headingType(node.Level)
		case *ast.Paragraph:
			elementType = "paragraph"
		case *ast.Link:
			elementType = "link"
		case *ast.Image:
			elementType = "image"
		case *ast.CodeBlock:
			elementType = "codeblock"
		case *ast.FencedCodeBlock:
			elementType = "codeblock"
		case *ast.CodeSpan:
			elementType = "code"
		case *ast.Blockquote:
			elementType = "blockquote"
		case *ast.List:
			elementType = "list"
		case *ast.ListItem:
			elementType = "listitem"
		default:
			return ast.WalkContinue, nil
		}

		// Only apply config classes if no inline class attribute exists
		// Inline attributes take precedence over config
		if _, exists := n.AttributeString("class"); !exists {
			classes := t.config.GetClasses(elementType, t.context)
			if classes != "" {
				n.SetAttributeString("class", classes)
			}
		}

		return ast.WalkContinue, nil
	})
}

func (t *Transformer) headingType(level int) string {
	switch level {
	case 1:
		return "heading1"
	case 2:
		return "heading2"
	case 3:
		return "heading3"
	case 4:
		return "heading4"
	case 5:
		return "heading5"
	case 6:
		return "heading6"
	default:
		return "heading1"
	}
}
