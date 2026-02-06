package navigation

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator checks that the generated HTML contains a <nav> element
// with links to all expected sections
type Validator struct {
	sections []string
}

// NewValidator creates a new navigation validator that will check
// for the presence of nav links to all given sections plus the home page
func NewValidator(sections []string) *Validator {
	return &Validator{
		sections: sections,
	}
}

// Validate checks the HTML content for a <nav> element containing links to all sections
func (v *Validator) Validate(htmlPath, buildDir string, content []byte) []error {
	var errs []error
	html := string(content)

	// Extract <nav> content
	navRegex := regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`)
	navMatch := navRegex.FindStringSubmatch(html)
	if len(navMatch) < 2 {
		errs = append(errs, fmt.Errorf("%s: missing <nav> element", htmlPath))
		return errs
	}

	navContent := navMatch[1]

	// Check home link — match href ending with /index.html" or starting with "index.html"
	// but not section/index.html
	homeHrefRegex := regexp.MustCompile(`href="(\.\./)*index\.html"`)
	if !strings.Contains(navContent, "Accueil") {
		errs = append(errs, fmt.Errorf("%s: navigation missing home link (Accueil)", htmlPath))
	}
	if !homeHrefRegex.MatchString(navContent) {
		errs = append(errs, fmt.Errorf("%s: navigation missing home href to index.html", htmlPath))
	}

	// Check each section link
	for _, section := range v.sections {
		expectedHref := section + "/index.html"
		if !strings.Contains(navContent, expectedHref) {
			errs = append(errs, fmt.Errorf("%s: navigation missing link to section %q (expected href containing %q)", htmlPath, section, expectedHref))
		}

		displayName := strings.ToUpper(section[:1]) + section[1:]
		if !strings.Contains(navContent, displayName) {
			errs = append(errs, fmt.Errorf("%s: navigation missing display name %q for section %q", htmlPath, displayName, section))
		}
	}

	return errs
}
