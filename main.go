package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed models/page.html
var pageTemplate string

// PageContext contient toutes les données disponibles pour les substitutions
type PageContext struct {
	Source      []byte // contenu markdown brut
	RelPath     string // chemin relatif du fichier source
	HTMLContent string // contenu HTML après conversion goldmark
}

// Substitution représente une substitution de template
type Substitution struct {
	Placeholder string                        // ex: "{{title}}"
	Resolve     func(ctx *PageContext) string // logique métier
}

// substitutions contient toutes les substitutions disponibles
var substitutions = []Substitution{
	{
		Placeholder: "{{title}}",
		Resolve: func(ctx *PageContext) string {
			return extractTitle(ctx.Source)
		},
	},
	{
		Placeholder: "{{content}}",
		Resolve: func(ctx *PageContext) string {
			return convertMdLinksToHtml(ctx.HTMLContent, ctx.RelPath)
		},
	},
}

// applySubstitutions applique toutes les substitutions au template
func applySubstitutions(template string, ctx *PageContext) string {
	result := template
	for _, sub := range substitutions {
		result = strings.Replace(result, sub.Placeholder, sub.Resolve(ctx), -1)
	}
	return result
}

// extractTitle extracts the first H1 heading from markdown content
func extractTitle(source []byte) string {
	re := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	match := re.FindSubmatch(source)
	if len(match) >= 2 {
		return string(match[1])
	}
	return "Untitled"
}

// resolveOutputPath determines the output path for a given source path
func resolveOutputPath(relPath string) string {
	switch {
	case relPath == "home.md":
		return "index.html"
	case relPath == "posts/index.md":
		return "posts/index.html"
	case strings.HasPrefix(relPath, "posts/"):
		filename := strings.TrimPrefix(relPath, "posts/")
		return "post/" + strings.TrimSuffix(filename, ".md") + ".html"
	default:
		return strings.TrimSuffix(relPath, ".md") + ".html"
	}
}

// convertMdLinksToHtml converts .md links to .html in the generated HTML
// sourceRelPath is the relative path of the source file (e.g., "posts/index.md")
func convertMdLinksToHtml(html string, sourceRelPath string) string {
	sourceDir := filepath.Dir(sourceRelPath)
	sourceOutputPath := resolveOutputPath(sourceRelPath)
	sourceOutputDir := filepath.Dir(sourceOutputPath)

	re := regexp.MustCompile(`href="([^"]*\.md)"`)
	return re.ReplaceAllStringFunc(html, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		linkPath := submatch[1]

		// Resolve the absolute path of the linked file relative to source
		var targetRelPath string
		if strings.HasPrefix(linkPath, "../") {
			// Go up one directory
			targetRelPath = filepath.Join(filepath.Dir(sourceDir), strings.TrimPrefix(linkPath, "../"))
		} else {
			targetRelPath = filepath.Join(sourceDir, linkPath)
		}
		targetRelPath = filepath.Clean(targetRelPath)

		// Get the output path for the target
		targetOutputPath := resolveOutputPath(targetRelPath)

		// Calculate relative path from source output to target output
		relLink, _ := filepath.Rel(sourceOutputDir, targetOutputPath)

		return fmt.Sprintf(`href="%s"`, relLink)
	})
}

func main() {
	// Process all markdown files in content/
	err := filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read markdown file
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Convert markdown to HTML using goldmark with GFM extensions
		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		)
		var buf bytes.Buffer
		if err := md.Convert(source, &buf); err != nil {
			return fmt.Errorf("converting %s: %w", path, err)
		}

		// Determine output path based on content type
		relPath, _ := filepath.Rel("content", path)
		outPath := filepath.Join("build", resolveOutputPath(relPath))

		// Build page context
		ctx := &PageContext{
			Source:      source,
			RelPath:     relPath,
			HTMLContent: buf.String(),
		}

		// Apply all substitutions
		html := applySubstitutions(pageTemplate, ctx)

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
	})

	if err != nil {
		panic(err)
	}
}
