package page

import (
	_ "embed"

	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page/substitution"
	"github.com/timtimjnvr/blog/internal/markdown"
	"github.com/timtimjnvr/blog/internal/styling"
)

//go:embed page.html
var defaultTemplate string

type Generator struct {
	htmlPageTemplate string
	markdownPath     string
	stylingConfig    styling.Config
	htmlOutputPath   string
	sectionName      string
	substitutions    substitution.Registry
}

func NewGenerator(markdownPath string, buildDir string, sectionName string, stylingConfig styling.Config) *Generator {
	var (
		baseName       = filepath.Base(markdownPath)
		outName        = strings.TrimSuffix(baseName, ".md") + ".html"
		htmlOutputPath = filepath.Join(buildDir, outName)
	)
	return &Generator{
		markdownPath:   markdownPath,
		htmlOutputPath: htmlOutputPath,
		stylingConfig:  stylingConfig,
	}
}

func (g *Generator) Generate() error {
	// Read markdown file
	markdDownSourceContent, err := os.ReadFile(g.markdownPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", g.markdownPath, err)
	}

	// Apply Styling convertions
	htmlContent, err := markdown.NewConverter(&g.stylingConfig, g.sectionName).Convert(markdDownSourceContent)
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %v", err)
	}

	// Project result inside the page template
	htmlContent, err = g.substitutions.Apply(g.htmlPageTemplate, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %v", err)
	}

	// Write HTML file
	if err := os.WriteFile(g.htmlOutputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", g.htmlOutputPath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", g.markdownPath, g.htmlOutputPath)
	return nil
}

func (g *Generator) Validate() error {
	return nil
}
