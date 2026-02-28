package page

import (
	_ "embed"

	"fmt"
	"path/filepath"

	"github.com/timtimjnvr/blog/internal/generator/page/filesystem"
	"github.com/timtimjnvr/blog/internal/generator/page/markdown"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/page/substitution"
	"github.com/timtimjnvr/blog/internal/generator/page/validation"
)

//go:embed page.html
var defaultTemplate string

type Generator struct {
	htmlPageTemplate    string
	sourceMDPath        string
	stylingConfig       styling.Config
	buildDir            string
	htmlContentBytes    []byte
	destinationHTMLPath string
	sectionName         string
	fs                  filesystem.FileSystem
	substitutions       *substitution.Registry
	validations         *validation.Registry
}

func NewGenerator(
	markdownSourcePath string,
	htmlOutputPath string,
	buildDir string,
	sectionName string,
	stylingConfig styling.Config,
	fs filesystem.FileSystem,
	substitutions *substitution.Registry,
	validations *validation.Registry,
) *Generator {
	return &Generator{
		htmlPageTemplate:    defaultTemplate,
		sourceMDPath:        markdownSourcePath,
		destinationHTMLPath: htmlOutputPath,
		buildDir:            buildDir,
		sectionName:         sectionName,
		stylingConfig:       stylingConfig,
		fs:                  fs,
		substitutions:       substitutions,
		validations:         validations,
	}
}

// Generate generates an html page by projecting the markdown file in the HTML template.
func (g *Generator) Generate() error {
	// Read markdown file
	markdDownSourceContent, err := g.fs.ReadFile(g.sourceMDPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", g.sourceMDPath, err)
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

	// Ensure output directory exists
	if err := g.fs.MkdirAll(filepath.Dir(g.destinationHTMLPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file
	htmlContentBytes := []byte(htmlContent)
	if err := g.fs.WriteFile(g.destinationHTMLPath, htmlContentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", g.destinationHTMLPath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", g.sourceMDPath, g.destinationHTMLPath)
	g.htmlContentBytes = htmlContentBytes
	return nil
}

func (g *Generator) Validate() error {
	return g.validations.Validate(g.destinationHTMLPath, g.buildDir, g.htmlContentBytes)
}
