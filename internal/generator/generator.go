package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/markdown"
	"github.com/timtimjnvr/blog/internal/substitution"
)

//go:embed page.html
var defaultTemplate string

// Generator handles site generation
type Generator struct {
	registry  *substitution.Registry[*context.PageContext]
	converter *markdown.Converter
	template  string
}

// New creates a new generator with the given substitution registry
func New(registry *substitution.Registry[*context.PageContext]) *Generator {
	return &Generator{
		registry:  registry,
		converter: markdown.NewConverter(),
		template:  defaultTemplate,
	}
}

// WithTemplate sets a custom template
func (g *Generator) WithTemplate(template string) *Generator {
	g.template = template
	return g
}

// Generate processes all markdown files from contentDir and outputs to buildDir
func (g *Generator) Generate(contentDir, buildDir string) error {
	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read markdown file
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Convert markdown to HTML
		htmlContent, err := g.converter.Convert(source)
		if err != nil {
			return fmt.Errorf("converting %s: %w", path, err)
		}

		// Determine output path
		relPath, _ := filepath.Rel(contentDir, path)
		outPath := filepath.Join(buildDir, markdown.ResolveOutputPath(relPath))

		// Build page context
		ctx := &context.PageContext{
			Source:      source,
			RelPath:     relPath,
			HTMLContent: htmlContent,
		}

		// Apply all substitutions
		html := g.registry.Apply(g.template, ctx)

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		// Write HTML file
		if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		fmt.Printf("Generated: %s -> %s\n", path, outPath)
		return nil
	})
}
