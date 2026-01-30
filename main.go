package main

import (
	"fmt"
	"os"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/generator"
	"github.com/timtimjnvr/blog/internal/styling"
	"github.com/timtimjnvr/blog/internal/substitution"
	"github.com/timtimjnvr/blog/internal/validator"
)

const (
	styleConfigPath = "styles/styles.json"
	scriptsDir      = "scripts"
)

func main() {
	// Create registry and register substitutions
	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.TitleSubstituter{})
	registry.Register(&substitution.ContentSubstituter{})

	// Load style configuration if it exists
	var styleConfig *styling.Config
	if _, err := os.Stat(styleConfigPath); err == nil {
		styleConfig, err = styling.LoadConfig(styleConfigPath)
		if err != nil {
			fmt.Printf("Error loading style configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loaded style configuration from %s\n", styleConfigPath)
	}

	// Generate site with validators
	gen := generator.New(registry).
		WithValidator(validator.NewImageValidator()).
		WithValidator(validator.NewScriptValidator()).
		WithValidator(validator.NewLinkValidator())

	if styleConfig != nil {
		gen = gen.WithStyleConfig(styleConfig)
	}

	// Add scripts directory if it exists
	if _, err := os.Stat(scriptsDir); err == nil {
		gen = gen.WithScriptsDir(scriptsDir)
	}

	if err := gen.Generate("content", "target/build"); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
