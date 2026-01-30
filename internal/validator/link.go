package validator

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LinkValidator checks that all links in HTML are accessible
type LinkValidator struct {
	// Timeout for HTTP requests to external links
	Timeout time.Duration
	// SkipExternal skips validation of external URLs
	SkipExternal bool
}

// NewLinkValidator creates a new link validator with default settings
func NewLinkValidator() *LinkValidator {
	return &LinkValidator{
		Timeout:      10 * time.Second,
		SkipExternal: false,
	}
}

// Validate checks all anchor href attributes in the HTML content
func (v *LinkValidator) Validate(htmlPath, buildDir string, content []byte) []ValidationError {
	var errors []ValidationError

	// Find all anchor href attributes
	linkRegex := regexp.MustCompile(`<a[^>]+href="([^"]+)"`)
	matches := linkRegex.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		href := string(match[1])

		// Skip fragment-only links (e.g., #section)
		if strings.HasPrefix(href, "#") {
			continue
		}

		// Skip mailto and tel links
		if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			continue
		}

		// Skip javascript links
		if strings.HasPrefix(href, "javascript:") {
			continue
		}

		if isExternalURL(href) {
			if v.SkipExternal {
				continue
			}
			if err := v.validateExternalLink(href); err != nil {
				errors = append(errors, ValidationError{
					File:    htmlPath,
					Message: fmt.Sprintf("external link not accessible: %s (%v)", href, err),
				})
			}
		} else {
			if err := v.validateLocalLink(href, htmlPath, buildDir); err != nil {
				errors = append(errors, ValidationError{
					File:    htmlPath,
					Message: fmt.Sprintf("local link not found: %s", href),
				})
			}
		}
	}

	return errors
}

// validateExternalLink checks if an external URL is accessible
func (v *LinkValidator) validateExternalLink(url string) error {
	client := &http.Client{
		Timeout: v.Timeout,
	}

	resp, err := client.Head(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// validateLocalLink checks if a local link target exists
func (v *LinkValidator) validateLocalLink(href, htmlPath, buildDir string) error {
	// Remove fragment identifier if present
	href = strings.Split(href, "#")[0]

	// Empty href after removing fragment means same-page link
	if href == "" {
		return nil
	}

	var linkPath string

	if strings.HasPrefix(href, "/") {
		// Absolute path from build root
		linkPath = filepath.Join(buildDir, href)
	} else {
		// Relative path from HTML file location
		htmlDir := filepath.Dir(htmlPath)
		linkPath = filepath.Join(htmlDir, href)
	}

	// Clean the path to resolve ../ etc
	linkPath = filepath.Clean(linkPath)

	// Check if path exists as-is (could be a file or directory)
	if _, err := os.Stat(linkPath); err == nil {
		return nil
	}

	// If path doesn't have an extension, check for index.html
	if filepath.Ext(linkPath) == "" {
		indexPath := filepath.Join(linkPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return nil
		}
	}

	return fmt.Errorf("path not found: %s", linkPath)
}
