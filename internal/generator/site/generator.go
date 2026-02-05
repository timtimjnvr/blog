package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(markdownPath, buildDir string) PageGenerator
)

type Generator struct {
	contentDir                string
	assetsDir                 string
	assetsOutDir              string
	buildDir                  string
	scriptsDir                string
	scriptsOutDir             string
	optionalStylingConfigPath string
	pageGeneratorFactory      pageGeneratorFactory
	sectionDirectoryNames     []string
	pagesGenerators           []PageGenerator
}

func NewGenerator() *Generator {
	return &Generator{
		contentDir:                "content/markdown",
		assetsDir:                 "content/assets",
		assetsOutDir:              "assets",
		buildDir:                  "target/build",
		scriptsDir:                "scripts",
		scriptsOutDir:             "scripts",
		optionalStylingConfigPath: "styles/styles.json",
		sectionDirectoryNames:     make([]string, 0),
		pagesGenerators:           make([]PageGenerator, 0),
		pageGeneratorFactory:      defaultPageGeneratorFactory,
	}
}

func defaultPageGeneratorFactory(markdownPath, buildDir string) PageGenerator {
	return page.NewGenerator(markdownPath, buildDir)
}

// WithPageGeneratorFactory sets a custom page generator factory.
func (g *Generator) WithPageGeneratorFactory(factory pageGeneratorFactory) *Generator {
	g.pageGeneratorFactory = factory
	return g
}

func (g *Generator) Generate() error {
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
	err := filepath.Walk(g.contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(path, ".md") {
			errs = append(errs, fmt.Errorf("wrong extension for file in %s", path))
			return nil
		}

		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(path, g.buildDir))
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
		return errors.Join(errs...)
	}

	return nil
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
