package script

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	jsparser "github.com/dop251/goja/parser"
)

// Validator checks that all scripts in HTML are accessible
type Validator struct{}

// NewValidator creates a new script validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks all script src attributes in the HTML content
func (v *Validator) Validate(htmlPath, buildDir string, content []byte) []error {
	var errs []error

	// Find all script src attributes
	scriptRegex := regexp.MustCompile(`<script[^>]+src="([^"]+)"`)
	matches := scriptRegex.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		src := string(match[1])

		// Skip external scripts
		if isExternalURL(src) {
			continue
		}

		scriptPath := v.resolveScriptPath(src, htmlPath, buildDir)

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("%s: local script not found: %s", htmlPath, src))
			continue
		}

		// Validate JavaScript syntax
		if syntaxErrs := v.validateJSSyntax(scriptPath); len(syntaxErrs) > 0 {
			errs = append(errs, syntaxErrs...)
		}
	}

	return errs
}

// isExternalURL checks if the URL is external (http/https)
func isExternalURL(src string) bool {
	return strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")
}

// resolveScriptPath resolves a script src to an absolute file path
func (v *Validator) resolveScriptPath(src, htmlPath, buildDir string) string {
	var scriptPath string

	if strings.HasPrefix(src, "/") {
		// Absolute path from build root
		scriptPath = filepath.Join(buildDir, src)
	} else {
		// Relative path from HTML file location
		htmlDir := filepath.Dir(htmlPath)
		scriptPath = filepath.Join(htmlDir, src)
	}

	// Clean the path to resolve ../ etc
	return filepath.Clean(scriptPath)
}

// validateJSSyntax parses the JavaScript file and returns any syntax errors
func (v *Validator) validateJSSyntax(scriptPath string) []error {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return []error{fmt.Errorf("%s: failed to read script: %v", scriptPath, err)}
	}

	_, parseErr := jsparser.ParseFile(nil, scriptPath, string(content), 0)
	if parseErr != nil {
		return []error{fmt.Errorf("%s: JavaScript syntax error: %v", scriptPath, parseErr)}
	}

	return nil
}
