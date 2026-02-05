package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(markdownPath, buildDir, section string, stylingConfig *styling.Config) PageGenerator
)

type Generator struct {
	contentDir                string
	assetsDir                 string
	assetsOutDir              string
	optionalStylingConfigPath string
	buildDir                  string
	scriptsDir                string
	scriptsOutDir             string
	pageGeneratorFactory      pageGeneratorFactory
	sectionDirectoryNames     []string
	pagesGenerators           []PageGenerator

	stylingConfig *styling.Config
}

func NewGenerator() *Generator {
	return &Generator{
		contentDir:                "content/markdown",
		assetsDir:                 "content/assets",
		assetsOutDir:              "assets",
		optionalStylingConfigPath: "styles/styles.json",
		buildDir:                  "target/build",
		scriptsDir:                "scripts",
		scriptsOutDir:             "scripts",
		sectionDirectoryNames:     make([]string, 0),
		pagesGenerators:           make([]PageGenerator, 0),
		pageGeneratorFactory:      defaultPageGeneratorFactory,
	}
}

func defaultPageGeneratorFactory(markdownPath, buildDir, section string, stylingConfig *styling.Config) PageGenerator {
	var config styling.Config
	if stylingConfig != nil {
		config = *stylingConfig
	}
	return page.NewGenerator(markdownPath, buildDir, section, config)
}

// WithPageGeneratorFactory sets a custom page generator factory.
func (g *Generator) WithPageGeneratorFactory(factory pageGeneratorFactory) *Generator {
	g.pageGeneratorFactory = factory
	return g
}

// WithContentDir sets the content directory.
func (g *Generator) WithContentDir(dir string) *Generator {
	g.contentDir = dir
	return g
}

// WithBuildDir sets the build output directory.
func (g *Generator) WithBuildDir(dir string) *Generator {
	g.buildDir = dir
	return g
}

// WithAssetsDir sets the assets directory.
func (g *Generator) WithAssetsDir(dir string) *Generator {
	g.assetsDir = dir
	return g
}

// WithScriptsDir sets the scripts directory.
func (g *Generator) WithScriptsDir(dir string) *Generator {
	g.scriptsDir = dir
	return g
}

// WithStylingConfigPath sets the styling configuration file path.
func (g *Generator) WithStylingConfigPath(path string) *Generator {
	g.optionalStylingConfigPath = path
	return g
}

func (g *Generator) Generate() error {
	if err := g.loadStylingConfig(); err != nil {
		return fmt.Errorf("failed to load styling configuration: %v", err)
	}

	if err := g.makeAllDirectories(); err != nil {
		return fmt.Errorf("failed to create output directories: %v", err)
	}

	if err := g.listSections(); err != nil {
		return fmt.Errorf("Failed to list site sections: %v", err)
	}

	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("Failed to copy assets: %v", err)
	}

	if err := g.copyScripts(); err != nil {
		return fmt.Errorf("Failed to copy scripts: %v", err)
	}

	if err := g.generatePages(); err != nil {
		return fmt.Errorf("Failed to generate pages: %v", err)
	}

	if err := g.Validate(); err != nil {
		return fmt.Errorf("Failed to validate site: %v", err)
	}
	return nil
}

func (g *Generator) loadStylingConfig() error {
	if _, err := os.Stat(g.optionalStylingConfigPath); err == nil {
		styleConfig, err := styling.LoadConfig(g.optionalStylingConfigPath)
		if err != nil {
			return fmt.Errorf("failed to LoadConfig: %v", err)
		}
		fmt.Printf("Loaded style configuration from %s\n", g.optionalStylingConfigPath)
		g.stylingConfig = styleConfig
	}
	return nil
}
func (g *Generator) listSections() error {
	return filepath.Walk(g.contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Listing sections by directory name
		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(g.contentDir, path)
		if err != nil {
			return err
		}

		// Sections are only top level directories of g.ContentDir
		if strings.Contains(relPath, "/") {
			return nil
		}

		// Section directory
		g.sectionDirectoryNames = append(g.sectionDirectoryNames, path)
		return nil
	})
}

func (g *Generator) makeAllDirectories() error {
	for _, d := range []string{g.assetsOutDir, g.scriptsOutDir, g.buildDir} {
		if err := os.MkdirAll(filepath.Dir(d), 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	return nil
}

func (g *Generator) copyAssets() error {
	return copyDir(g.assetsDir, filepath.Join(g.buildDir, "assets"), nil)
}

func (g *Generator) copyScripts() error {
	return copyDir(g.scriptsDir, filepath.Join(g.buildDir, "scripts"), func(path string) bool {
		return strings.HasSuffix(path, ".js")
	})
}

func (g *Generator) generatePages() error {
	errs := make([]error, 0)
	err := filepath.Walk(g.contentDir, func(pageFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(pageFilePath, ".md") {
			errs = append(errs, fmt.Errorf("wrong extension for file in %s", pageFilePath))
			return nil
		}

		// Page section is the directory between content dir and file name
		pageSection, err := extractSection(g.contentDir, pageFilePath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(pageFilePath, g.buildDir, pageSection, g.stylingConfig))
		return nil
	})

	if err != nil {
		errs = append(errs, err)
	}

	for _, generator := range g.pagesGenerators {
		if err := generator.Generate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		// Empty the page generators
		g.pagesGenerators = make([]PageGenerator, 0)
		return errors.Join(errs...)
	}

	return nil
}

// extractSection returns the section (subdirectory path) for a file relative to the content directory.
// For example, if contentDir is "content/markdown" and filePath is "content/markdown/blog/post.md",
// the section returned is "blog". Returns empty string if the file is at the root of contentDir.
func extractSection(contentDir, filePath string) (string, error) {
	relPath, err := filepath.Rel(contentDir, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
	}
	section := filepath.Dir(relPath)
	if section == "." {
		return "", nil
	}
	return section, nil
}

func (g *Generator) Validate() error {
	errs := make([]error, 0)
	for _, pg := range g.pagesGenerators {
		if err := pg.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
