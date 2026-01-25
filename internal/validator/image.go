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

// ImageValidator checks that all images in HTML are accessible
type ImageValidator struct {
	// Timeout for HTTP requests to external images
	Timeout time.Duration
	// SkipExternal skips validation of external URLs
	SkipExternal bool
}

// NewImageValidator creates a new image validator with default settings
func NewImageValidator() *ImageValidator {
	return &ImageValidator{
		Timeout:      10 * time.Second,
		SkipExternal: false,
	}
}

// Validate checks all img src attributes in the HTML content
func (v *ImageValidator) Validate(htmlPath, buildDir string, content []byte) []ValidationError {
	var errors []ValidationError

	// Find all img src attributes
	imgRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	matches := imgRegex.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		src := string(match[1])

		if isExternalURL(src) {
			if v.SkipExternal {
				continue
			}
			if err := v.validateExternalImage(src); err != nil {
				errors = append(errors, ValidationError{
					File:    htmlPath,
					Message: fmt.Sprintf("external image not accessible: %s (%v)", src, err),
				})
			}
		} else {
			if err := v.validateLocalImage(src, htmlPath, buildDir); err != nil {
				errors = append(errors, ValidationError{
					File:    htmlPath,
					Message: fmt.Sprintf("local image not found: %s", src),
				})
			}
		}
	}

	return errors
}

// isExternalURL checks if the URL is external (http/https)
func isExternalURL(src string) bool {
	return strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")
}

// validateExternalImage checks if an external image URL is accessible
func (v *ImageValidator) validateExternalImage(url string) error {
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

// validateLocalImage checks if a local image file exists
func (v *ImageValidator) validateLocalImage(src, htmlPath, buildDir string) error {
	var imagePath string

	if strings.HasPrefix(src, "/") {
		// Absolute path from build root
		imagePath = filepath.Join(buildDir, src)
	} else {
		// Relative path from HTML file location
		htmlDir := filepath.Dir(htmlPath)
		imagePath = filepath.Join(htmlDir, src)
	}

	// Clean the path to resolve ../ etc
	imagePath = filepath.Clean(imagePath)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return err
	}

	return nil
}
