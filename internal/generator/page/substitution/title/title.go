package title

import (
	"fmt"
	"regexp"
)

// Substituter resolves {{title}} placeholder
type Substituter struct {
}

func NewSubstituer() Substituter {
	return Substituter{}
}

func (t Substituter) Placeholder() string {
	return "{{title}}"
}

func (t Substituter) Resolve(content string) (string, error) {
	re := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	match := re.FindSubmatch([]byte(content))
	if len(match) >= 2 {
		return string(match[1]), nil
	}

	return "", fmt.Errorf("Could not find a page title")
}
