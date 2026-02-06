package navigation

import (
	"fmt"
	"strings"
)

// Substituter resolves {{navigation}} placeholder with an auto-generated nav bar
type Substituter struct {
	sections       []string
	currentSection string
}

func NewSubstituer(sections []string, currentSection string) Substituter {
	return Substituter{
		sections:       sections,
		currentSection: currentSection,
	}
}

func (n Substituter) Placeholder() string {
	return "{{navigation}}"
}

func (n Substituter) Resolve(_ string) (string, error) {
	prefix := relativePrefix(n.currentSection)

	var links []string
	links = append(links, fmt.Sprintf(`<a href="%sindex.html" class="hover:underline">Accueil</a>`, prefix))

	for _, section := range n.sections {
		href := prefix + section + "/index.html"
		displayName := strings.ToUpper(section[:1]) + section[1:]
		links = append(links, fmt.Sprintf(`<a href="%s" class="hover:underline">%s</a>`, href, displayName))
	}

	return fmt.Sprintf(`<nav class="flex gap-4">%s</nav>`, strings.Join(links, "\n    ")), nil
}

// relativePrefix returns the "../" prefix needed to reach the site root from the current section.
func relativePrefix(currentSection string) string {
	if currentSection == "" {
		return ""
	}
	depth := strings.Count(currentSection, "/") + 1
	return strings.Repeat("../", depth)
}
