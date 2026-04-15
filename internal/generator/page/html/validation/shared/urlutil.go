// Package shared provides utilities shared across HTML validators.
package shared

import (
	"path/filepath"
	"strings"
)

// IsExternalURL reports whether src is an external URL (http or https).
func IsExternalURL(src string) bool {
	return strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")
}

// ResolveLocalPath resolves src to an absolute file path relative to the HTML
// file or the build root, matching how browsers resolve relative vs absolute paths.
//
// Absolute paths (starting with "/") are resolved from buildDir.
// Relative paths are resolved from the directory containing htmlPath.
// The result is cleaned to resolve any ".." components.
func ResolveLocalPath(src, htmlPath, buildDir string) string {
	var resolved string
	if strings.HasPrefix(src, "/") {
		resolved = filepath.Join(buildDir, src)
	} else {
		resolved = filepath.Join(filepath.Dir(htmlPath), src)
	}
	return filepath.Clean(resolved)
}
