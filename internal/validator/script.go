package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	jsparser "github.com/dop251/goja/parser"
)

// ScriptValidator checks that all scripts in HTML are accessible
type ScriptValidator struct{}

// NewScriptValidator creates a new script validator
func NewScriptValidator() *ScriptValidator {
	return &ScriptValidator{}
}

// Validate checks all script src attributes in the HTML content
func (v *ScriptValidator) Validate(htmlPath, buildDir string, content []byte) []ValidationError {
	var errors []ValidationError

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
			errors = append(errors, ValidationError{
				File:    htmlPath,
				Message: fmt.Sprintf("local script not found: %s", src),
			})
			continue
		}

		// Validate JavaScript syntax
		if syntaxErrs := v.validateJSSyntax(scriptPath); len(syntaxErrs) > 0 {
			errors = append(errors, syntaxErrs...)
		}
	}

	return errors
}

// resolveScriptPath resolves a script src to an absolute file path
func (v *ScriptValidator) resolveScriptPath(src, htmlPath, buildDir string) string {
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
func (v *ScriptValidator) validateJSSyntax(scriptPath string) []ValidationError {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return []ValidationError{{
			File:    scriptPath,
			Message: fmt.Sprintf("failed to read script: %v", err),
		}}
	}

	_, parseErr := jsparser.ParseFile(nil, scriptPath, string(content), 0)
	if parseErr != nil {
		return []ValidationError{{
			File:    scriptPath,
			Message: fmt.Sprintf("JavaScript syntax error: %v", parseErr),
		}}
	}

	return nil
}
