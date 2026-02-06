package page

import (
	_ "embed"

	"fmt"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page/filesystem"
	"github.com/timtimjnvr/blog/internal/generator/page/markdown"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/page/substitution"
	"github.com/timtimjnvr/blog/internal/generator/page/validation"
)

//go:embed page.html
var defaultTemplate string

type Generator struct {
	htmlPageTemplate string
	markdownPath     string
	stylingConfig    styling.Config
	buildDir         string
	htmlContentBytes []byte
	htmlOutputPath   string
	sectionName      string
	fs               filesystem.FileSystem
	substitutions    *substitution.Registry
	validations      *validation.Registry
}

func NewGenerator(
	markdownPath string,
	buildDir string,
	sectionName string,
	stylingConfig styling.Config,
	fs filesystem.FileSystem,
	substitutions *substitution.Registry,
	validations *validation.Registry,
) *Generator {
	var (
		baseName       = filepath.Base(markdownPath)
		outName        = strings.TrimSuffix(baseName, ".md") + ".html"
		htmlOutputPath string
	)
	// Preserve directory structure: section pages go into section subdirectory
	if sectionName != "" {
		htmlOutputPath = filepath.Join(buildDir, sectionName, outName)
	} else {
		htmlOutputPath = filepath.Join(buildDir, outName)
	}
	return &Generator{
		htmlPageTemplate: defaultTemplate,
		markdownPath:     markdownPath,
		htmlOutputPath:   htmlOutputPath,
		buildDir:         buildDir,
		sectionName:      sectionName,
		stylingConfig:    stylingConfig,
		fs:               fs,
		substitutions:    substitutions,
		validations:      validations,
	}
}

func (g *Generator) Generate() error {
	// Read markdown file
	markdDownSourceContent, err := g.fs.ReadFile(g.markdownPath)
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

	// Ensure output directory exists
	if err := g.fs.MkdirAll(filepath.Dir(g.htmlOutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file
	htmlContentBytes := []byte(htmlContent)
	if err := g.fs.WriteFile(g.htmlOutputPath, htmlContentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", g.htmlOutputPath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", g.markdownPath, g.htmlOutputPath)
	g.htmlContentBytes = htmlContentBytes
	return nil
}

func (g *Generator) Validate() error {
	return g.validations.Validate(g.htmlOutputPath, g.buildDir, g.htmlContentBytes)
}
