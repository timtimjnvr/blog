package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timtimjnvr/blog/internal/generator/page"
	"github.com/timtimjnvr/blog/internal/generator/page/filesystem"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/page/substitution"
	"github.com/timtimjnvr/blog/internal/generator/page/validation"
	"github.com/timtimjnvr/blog/internal/generator/section"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, stylingConfig *styling.Config, sections []section.Section) PageGenerator
)

// Generator is the site generator which allows to generate and validate the site
// All files and directories attributes are relative to the project root.
type Generator struct {
	contentDir                string
	assetsDir                 string
	assetsOutDir              string
	optionalStylingConfigPath string
	buildDir                  string
	scriptsDir                string
	scriptsOutDir             string
	pageGeneratorFactory      pageGeneratorFactory
	sections                  []section.Section
	pagesGenerators           []PageGenerator
	stylingConfig             *styling.Config
}

func NewGenerator() (*Generator, error) {
	g := &Generator{
		contentDir:                "./content/markdown",
		buildDir:                  "./target/build",
		assetsDir:                 "./content/assets",
		assetsOutDir:              "./target/build/assets",
		scriptsDir:                "./scripts",
		scriptsOutDir:             "./target/build/scripts",
		optionalStylingConfigPath: "./styles/styles.json",
		sections:                  make([]section.Section, 0),
		pagesGenerators:           make([]PageGenerator, 0),
		pageGeneratorFactory:      defaultPageGeneratorFactory,
	}

	return g, nil
}

func (g *Generator) Generate() error {
	if err := g.loadStylingConfig(); err != nil {
		return fmt.Errorf("failed to load styling configuration: %w", err)
	}

	if err := g.makeAllDirectories(); err != nil {
		return fmt.Errorf("failed to create output directories: %w", err)
	}

	if err := g.listSections(); err != nil {
		return fmt.Errorf("failed to list site sections: %w", err)
	}

	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	if err := g.copyScripts(); err != nil {
		return fmt.Errorf("failed to copy scripts: %w", err)
	}

	if err := g.generatePages(); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
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

func defaultPageGeneratorFactory(sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, stylingConfig *styling.Config, sections []section.Section) PageGenerator {
	var config styling.Config
	if stylingConfig != nil {
		config = *stylingConfig
	}

	var (
		fs            = filesystem.NewOSFileSystem()
		substitutions = substitution.NewRegistry(destinationHTMLPath, sourceMDPath, assetsPathTranslater, linksPathTranslater, sections, pageSection)
		validations   = validation.NewRegistry(sections)
	)

	return page.NewGenerator(sourceMDPath, destinationHTMLPath, buildDir, pageSection, config, fs, substitutions, validations)
}

func (g *Generator) loadStylingConfig() error {
	if _, err := os.Stat(g.optionalStylingConfigPath); err == nil {
		styleConfig, err := styling.LoadConfig(g.optionalStylingConfigPath)
		if err != nil {
			return fmt.Errorf("failed to LoadConfig: %w", err)
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

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(g.contentDir, path)
		if err != nil {
			return err
		}

		// Only process top-level directories (and the root itself)
		if strings.Contains(relPath, "/") {
			return nil
		}

		// Root content directory: home section has an empty DirName
		if relPath == "." {
			displayName := extractSectionTitle(filepath.Join(path, "index.md"), "home")
			g.sections = append(g.sections, section.Section{
				DirName:     "",
				DisplayName: displayName,
			})
			return nil
		}

		// Sub-section: read display name from # title in section's index.md
		displayName := extractSectionTitle(filepath.Join(path, "index.md"), relPath)
		g.sections = append(g.sections, section.Section{
			DirName:     relPath,
			DisplayName: displayName,
		})
		return nil
	})
}

// extractSectionTitle reads the # title from the index.md of a section directory.
// Falls back to the capitalized dirName if no index.md or no title is found.
func extractSectionTitle(indexMDPath, dirName string) string {
	content, err := os.ReadFile(indexMDPath)
	if err != nil {
		return strings.ToUpper(dirName[:1]) + dirName[1:]
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return strings.ToUpper(dirName[:1]) + dirName[1:]
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
	assetsPathTranslater := NewPathResolver(g.assetsDir, filepath.Join(g.buildDir, "assets"))
	linksPathTranslater := NewPathResolver(g.contentDir, g.buildDir)

	errs := make([]error, 0)
	err := filepath.Walk(g.contentDir, func(markDownFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(markDownFilePath, ".md") {
			errs = append(errs, fmt.Errorf("wrong extension for file in %s", markDownFilePath))
			return nil
		}

		// Page section is the directory between content dir and file name
		pageSection, err := extractSection(g.contentDir, markDownFilePath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		pageFilePathRelToContentDir, err := filepath.Rel(g.contentDir, markDownFilePath)
		if err != nil {
			return fmt.Errorf("cannot compute relative path of %s from %s: %w", markDownFilePath, g.contentDir, err)
		}

		htmlOutputPath := filepath.Join(g.buildDir, strings.TrimSuffix(pageFilePathRelToContentDir, ".md")+".html")
		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(markDownFilePath, htmlOutputPath, g.buildDir, pageSection, assetsPathTranslater, linksPathTranslater, g.stylingConfig, g.sections))
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
