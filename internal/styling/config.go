package styling

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// validElementTypes defines all recognized element type keys.
var validElementTypes = map[string]bool{
	"heading1":   true,
	"heading2":   true,
	"heading3":   true,
	"heading4":   true,
	"heading5":   true,
	"heading6":   true,
	"paragraph":  true,
	"link":       true,
	"image":      true,
	"codeblock":  true,
	"code":       true,
	"blockquote": true,
	"list":       true,
	"listitem":   true,
}

// Config holds CSS class mappings for HTML elements.
// Classes defined here are added to elements during Markdown conversion.
type Config struct {
	// Elements maps element types to CSS classes.
	// Supported keys: heading1-heading6, paragraph, link, image, codeblock, code, blockquote, list, listitem
	Elements map[string]string `json:"elements"`

	// Contexts allows different styling per page context (e.g., "post", "index").
	// The key is the context name, value is an element-to-classes map.
	Contexts map[string]map[string]string `json:"contexts"`
}

// NewConfig creates a Config with empty maps.
func NewConfig() Config {
	return Config{
		Elements: make(map[string]string),
		Contexts: make(map[string]map[string]string),
	}
}

// LoadConfig loads a Config from a JSON file.
// Returns an error if the file cannot be read, parsed, or contains invalid keys.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading style config: %w", err)
	}

	return ParseConfig(data)
}

// ParseConfig parses JSON data into a Config.
// Returns an error if the data is invalid or contains unrecognized element types.
func ParseConfig(data []byte) (*Config, error) {
	config := NewConfig()

	if len(data) == 0 {
		return config, nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing style config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks that all element type keys in the config are recognized.
// Returns an error listing all invalid keys if any are found.
func (c *Config) Validate() error {
	var invalidKeys []string

	// Check global elements
	for key := range c.Elements {
		if !validElementTypes[key] {
			invalidKeys = append(invalidKeys, fmt.Sprintf("elements.%s", key))
		}
	}

	// Check context-specific elements
	for contextName, contextMap := range c.Contexts {
		for key := range contextMap {
			if !validElementTypes[key] {
				invalidKeys = append(invalidKeys, fmt.Sprintf("contexts.%s.%s", contextName, key))
			}
		}
	}

	if len(invalidKeys) > 0 {
		return fmt.Errorf("invalid style config keys: %s. Valid keys are: %s",
			strings.Join(invalidKeys, ", "),
			strings.Join(validElementTypesList(), ", "))
	}

	return nil
}

// validElementTypesList returns a sorted list of valid element types.
func validElementTypesList() []string {
	return []string{
		"heading1", "heading2", "heading3", "heading4", "heading5", "heading6",
		"paragraph", "link", "image", "codeblock", "code", "blockquote", "list", "listitem",
	}
}

// GetClasses returns the CSS classes for a given element type.
// If a context is provided and has a specific override, it takes precedence.
func (c *Config) GetClasses(elementType, context string) string {
	// Check context-specific override first
	if context != "" {
		if contextMap, ok := c.Contexts[context]; ok {
			if classes, ok := contextMap[elementType]; ok {
				return classes
			}
		}
	}

	// Fall back to global element classes
	if classes, ok := c.Elements[elementType]; ok {
		return classes
	}

	return ""
}
