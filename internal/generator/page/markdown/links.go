package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ConvertAssetPaths adjusts relative asset paths (images) for the output structure.
// Since markdown files are in content/markdown/ but output structure flattens this,
// relative paths that go outside content/markdown/ need adjustment.
// sourceRelPath is the relative path from content/markdown/ (e.g., "posts/second-post.md")
func ConvertAssetPaths(html string, sourceRelPath string) string {
	sourceDir := filepath.Dir(sourceRelPath)
	sourceOutputDir := filepath.Dir(ResolveOutputPath(sourceRelPath))

	// Match img src attributes with relative paths
	re := regexp.MustCompile(`(<img[^>]+src=")([^"]+)(")`)
	return re.ReplaceAllStringFunc(html, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 4 {
			return match
		}

		prefix := submatch[1]
		src := submatch[2]
		suffix := submatch[3]

		// Skip external URLs and absolute paths
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "/") {
			return match
		}

		// Count how many levels up the path goes
		upCount := 0
		remaining := src
		for strings.HasPrefix(remaining, "../") {
			upCount++
			remaining = strings.TrimPrefix(remaining, "../")
		}

		if upCount == 0 {
			return match
		}

		// Calculate the target directory from source perspective
		targetDir := sourceDir
		for i := 0; i < upCount; i++ {
			targetDir = filepath.Dir(targetDir)
		}
		targetPath := filepath.Join(targetDir, remaining)
		targetPath = filepath.Clean(targetPath)

		// Remove "content/" prefix if present since assets are copied without it
		targetPath = strings.TrimPrefix(targetPath, "content/")
		if after, ok := strings.CutPrefix(targetPath, "../"); ok {
			targetPath = after
		}

		// Calculate relative path from output directory
		relPath, _ := filepath.Rel(sourceOutputDir, targetPath)

		return prefix + relPath + suffix
	})
}

// ResolveOutputPath determines the output path for a given source path
// Simply replaces .md extension with .html, preserving the directory structure
func ResolveOutputPath(relPath string) string {
	return strings.TrimSuffix(relPath, ".md") + ".html"
}

// ConvertMdLinksToHtml converts .md links to .html in the generated HTML
// sourceRelPath is the relative path of the source file (e.g., "posts/index.md")
// If sourceRelPath is empty, simply replaces .md with .html in all links
func ConvertMdLinksToHtml(html string, sourceRelPath string) string {
	re := regexp.MustCompile(`href="([^"]*\.md)"`)

	// Simple mode: just replace .md with .html when no source path context
	if sourceRelPath == "" {
		return re.ReplaceAllStringFunc(html, func(match string) string {
			submatch := re.FindStringSubmatch(match)
			if len(submatch) < 2 {
				return match
			}
			linkPath := submatch[1]
			htmlPath := strings.TrimSuffix(linkPath, ".md") + ".html"
			return fmt.Sprintf(`href="%s"`, htmlPath)
		})
	}

	sourceDir := filepath.Dir(sourceRelPath)
	sourceOutputPath := ResolveOutputPath(sourceRelPath)
	sourceOutputDir := filepath.Dir(sourceOutputPath)

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
