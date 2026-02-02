package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/markdown"
	"github.com/timtimjnvr/blog/internal/styling"
	"github.com/timtimjnvr/blog/internal/substitution"
	"github.com/timtimjnvr/blog/internal/validator"
)

//go:embed page.html
var defaultTemplate string

// LegacyGenerator handles site generation
type LegacyGenerator struct {
	registry    *substitution.Registry[*context.PageContext]
	styleConfig *styling.Config
	template    string
	validators  *validator.Runner
	scriptsDir  string
	assetsDir   string
}

// New creates a new generator with the given substitution registry
func New(registry *substitution.Registry[*context.PageContext]) *LegacyGenerator {
	return &LegacyGenerator{
		registry:    registry,
		styleConfig: nil,
		template:    defaultTemplate,
		validators:  validator.NewRunner(),
	}
}

// WithStyleConfig sets a style configuration for CSS class injection
func (g *LegacyGenerator) WithStyleConfig(config *styling.Config) *LegacyGenerator {
	g.styleConfig = config
	return g
}

// WithValidator adds a validator to the generator
func (g *LegacyGenerator) WithValidator(v validator.Validator) *LegacyGenerator {
	g.validators.Register(v)
	return g
}

// WithTemplate sets a custom template
func (g *LegacyGenerator) WithTemplate(template string) *LegacyGenerator {
	g.template = template
	return g
}

// WithScriptsDir sets a directory containing JavaScript files to copy
func (g *LegacyGenerator) WithScriptsDir(dir string) *LegacyGenerator {
	g.scriptsDir = dir
	return g
}

// WithAssetsDir sets a directory containing static assets to copy
func (g *LegacyGenerator) WithAssetsDir(dir string) *LegacyGenerator {
	g.assetsDir = dir
	return g
}

// Generate processes all markdown files from contentDir and outputs to buildDir
func (g *LegacyGenerator) Generate(contentDir, buildDir string) error {
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

	// Copy scripts if scriptsDir is set
	if g.scriptsDir != "" {
		if err := g.copyScripts(buildDir); err != nil {
			return err
		}
	}

	// Copy assets if assetsDir is set
	if g.assetsDir != "" {
		if err := g.copyAssets(buildDir); err != nil {
			return err
		}
	}

	// Run validators on generated HTML files
	return g.validate(generatedFiles, buildDir)
}

// validate runs all validators on the generated HTML files
func (g *LegacyGenerator) validate(htmlFiles []string, buildDir string) error {
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
func (g *LegacyGenerator) processMarkdown(path, relPath, buildDir string) error {
	// Read markdown file
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	// Determine context from path (e.g., "post" for files in posts/)
	pageContext := g.deriveContext(relPath)

	// Create converter with styling config and context
	converter := markdown.NewConverter(g.styleConfig, pageContext)

	// Convert markdown to HTML
	htmlContent, err := converter.Convert(source)
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

// deriveContext determines the styling context from the file path.
// For example, files in "posts/" get context "post".
func (g *LegacyGenerator) deriveContext(relPath string) string {
	if strings.HasPrefix(relPath, "posts/") {
		return "post"
	}
	return ""
}

// copyAsset copies a static file to the build directory
func (g *LegacyGenerator) copyAsset(path, relPath, buildDir string) error {
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

// copyScripts copies JavaScript files from scriptsDir to buildDir/scripts
func (g *LegacyGenerator) copyScripts(buildDir string) error {
	scriptsOutDir := filepath.Join(buildDir, "scripts")

	return filepath.Walk(g.scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only copy .js files
		if !strings.HasSuffix(path, ".js") {
			return nil
		}

		relPath, _ := filepath.Rel(g.scriptsDir, path)
		outPath := filepath.Join(scriptsOutDir, relPath)

		// Read source file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading script %s: %w", path, err)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		// Write file
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("writing script %s: %w", outPath, err)
		}

		fmt.Printf("Copied script: %s -> %s\n", path, outPath)
		return nil
	})
}

// copyAssets copies static assets from assetsDir to buildDir/assets
func (g *LegacyGenerator) copyAssets(buildDir string) error {
	// Assets are referenced from markdown relative to content/markdown/,
	// so strip "content/" prefix to get correct output path
	assetsOutputDir := strings.TrimPrefix(g.assetsDir, "content/")
	assetsOutDir := filepath.Join(buildDir, assetsOutputDir)

	return filepath.Walk(g.assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(g.assetsDir, path)
		outPath := filepath.Join(assetsOutDir, relPath)

		// Read source file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading asset %s: %w", path, err)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		// Write file
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("writing asset %s: %w", outPath, err)
		}

		fmt.Printf("Copied asset: %s -> %s\n", path, outPath)
		return nil
	})
}
