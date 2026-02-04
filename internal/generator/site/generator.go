package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page"
)

type PageGenerator interface {
	Generate() error
}

type PageGeneratorFactory func(markdownPath, buildDir string) PageGenerator

type Generator struct {
	ContentDir            string
	AssetsDir             string
	AssetsOutDir          string
	BuildDir              string
	ScriptsDir            string
	ScriptsOutDir         string
	SectionDirectoryNames []string
	GeneratedPagesPath    []string
	pageGeneratorFactory  PageGeneratorFactory
}

func New() *Generator {
	return &Generator{
		ContentDir:            "content/markdown",
		AssetsDir:             "content/assets",
		AssetsOutDir:          "assets",
		BuildDir:              "target/build",
		ScriptsDir:            "scripts",
		ScriptsOutDir:         "scripts",
		SectionDirectoryNames: make([]string, 0),
		pageGeneratorFactory:  defaultPageGeneratorFactory,
	}
}

func defaultPageGeneratorFactory(markdownPath, buildDir string) PageGenerator {
	return page.New(markdownPath, buildDir)
}

// WithPageGeneratorFactory sets a custom page generator factory.
func (g *Generator) WithPageGeneratorFactory(factory PageGeneratorFactory) *Generator {
	g.pageGeneratorFactory = factory
	return g
}

func (g *Generator) Generate() error {
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

	if err := g.validate(); err != nil {
		return fmt.Errorf("Failed to validate site: %v", err)
	}
	return nil
}

func (g *Generator) copyAssets() error {
	return copyDir(g.AssetsDir, filepath.Join(g.BuildDir, "assets"), nil)
}

func (g *Generator) copyScripts() error {
	return copyDir(g.ScriptsDir, filepath.Join(g.BuildDir, "scripts"), func(path string) bool {
		return strings.HasSuffix(path, ".js")
	})
}

func (g *Generator) listSections() error {
	return filepath.Walk(g.ContentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Listing sections by directory name
		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(g.ContentDir, path)
		if err != nil {
			return err
		}

		// Sections are only top level directories of g.ContentDir
		if strings.Contains(relPath, "/") {
			return nil
		}

		// Section directory
		g.SectionDirectoryNames = append(g.SectionDirectoryNames, path)
		return nil
	})
}

func (g *Generator) generatePages() error {
	errs := make([]error, 0)
	g.GeneratedPagesPath = make([]string, 0)

	err := filepath.Walk(g.ContentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(path, ".md") {
			return fmt.Errorf("wrong extension for file in %s", path)
		}

		pageGenerator := g.pageGeneratorFactory(path, g.BuildDir)
		if err := pageGenerator.Generate(); err != nil {
			errs = append(errs, err)
			return nil
		}

		g.GeneratedPagesPath = append(g.GeneratedPagesPath, path)

		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (g *Generator) validate() error {
	return nil
}
