package site

import (
	"fmt"

	"github.com/timtimjnvr/blog/internal/generator"
)

type Generator struct {
	ContentDir string
	AssetsDir  string
	BuildDir   string
	ScriptsDir string

	PageGenerator generator.Generator
}

func New(PageGenerator generator.Generator) Generator {
	return Generator{
		ContentDir: "content/markdown",
		AssetsDir:  "content/assets",
		BuildDir:   "target/build",
		ScriptsDir: "scripts",
	}
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
	return nil
}

func (g *Generator) copyScripts() error {
	return nil
}

func (g *Generator) listSections() error {
	return nil
}

func (g *Generator) generatePages() error {
	return nil
}

func (g *Generator) validate() error {
	return nil
}
