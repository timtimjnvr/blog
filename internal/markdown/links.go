package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ResolveOutputPath determines the output path for a given source path
// Simply replaces .md extension with .html, preserving the directory structure
func ResolveOutputPath(relPath string) string {
	return strings.TrimSuffix(relPath, ".md") + ".html"
}

// ConvertMdLinksToHtml converts .md links to .html in the generated HTML
// sourceRelPath is the relative path of the source file (e.g., "posts/index.md")
func ConvertMdLinksToHtml(html string, sourceRelPath string) string {
	sourceDir := filepath.Dir(sourceRelPath)
	sourceOutputPath := ResolveOutputPath(sourceRelPath)
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
		if after, ok := strings.CutPrefix(linkPath, "../"); ok {
			// Go up one directory
			targetRelPath = filepath.Join(filepath.Dir(sourceDir), after)
		} else {
			targetRelPath = filepath.Join(sourceDir, linkPath)
		}
		targetRelPath = filepath.Clean(targetRelPath)

		// Get the output path for the target
		targetOutputPath := ResolveOutputPath(targetRelPath)

		// Calculate relative path from source output to target output
		relLink, _ := filepath.Rel(sourceOutputDir, targetOutputPath)

		return fmt.Sprintf(`href="%s"`, relLink)
	})
}
