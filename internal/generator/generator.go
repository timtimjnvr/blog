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
	"github.com/timtimjnvr/blog/internal/validator"
)

//go:embed page.html
var defaultTemplate string

// Generator handles site generation
type Generator struct {
	registry   *substitution.Registry[*context.PageContext]
	converter  *markdown.Converter
	template   string
	validators *validator.Runner
}

// New creates a new generator with the given substitution registry
func New(registry *substitution.Registry[*context.PageContext]) *Generator {
	return &Generator{
		registry:   registry,
		converter:  markdown.NewConverter(),
		template:   defaultTemplate,
		validators: validator.NewRunner(),
	}
}

// WithValidator adds a validator to the generator
func (g *Generator) WithValidator(v validator.Validator) *Generator {
	g.validators.Register(v)
	return g
}

// WithTemplate sets a custom template
func (g *Generator) WithTemplate(template string) *Generator {
	g.template = template
	return g
}

// Generate processes all markdown files from contentDir and outputs to buildDir
func (g *Generator) Generate(contentDir, buildDir string) error {
	var generatedFiles []string

	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)

		// Handle markdown files
		if strings.HasSuffix(path, ".md") {
			outPath := filepath.Join(buildDir, markdown.ResolveOutputPath(relPath))
			if err := g.processMarkdown(path, relPath, buildDir); err != nil {
				return err
			}
			generatedFiles = append(generatedFiles, outPath)
			return nil
		}

		// Copy static assets (images, etc.)
		return g.copyAsset(path, relPath, buildDir)
	})

	if err != nil {
		return err
	}

	// Run validators on generated HTML files
	return g.validate(generatedFiles, buildDir)
}

// validate runs all validators on the generated HTML files
func (g *Generator) validate(htmlFiles []string, buildDir string) error {
	allResults := &validator.ValidationResult{}

	for _, htmlPath := range htmlFiles {
		content, err := os.ReadFile(htmlPath)
		if err != nil {
			return fmt.Errorf("reading %s for validation: %w", htmlPath, err)
		}

		result := g.validators.Validate(htmlPath, buildDir, content)
		allResults.Add(result.Errors...)
	}

	if allResults.HasErrors() {
		fmt.Println("\nValidation errors:")
		for _, err := range allResults.Errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("validation failed with %d error(s)", len(allResults.Errors))
	}

	return nil
}

// processMarkdown converts a markdown file to HTML
func (g *Generator) processMarkdown(path, relPath, buildDir string) error {
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
}

// copyAsset copies a static file to the build directory
func (g *Generator) copyAsset(path, relPath, buildDir string) error {
	outPath := filepath.Join(buildDir, relPath)

	// Read source file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", outPath, err)
	}

	// Write file
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("Copied: %s -> %s\n", path, outPath)
	return nil
}
