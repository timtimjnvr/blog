package main

import (
	"fmt"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/generator"
	"github.com/timtimjnvr/blog/internal/substitution"
	"github.com/timtimjnvr/blog/internal/validator"
)

func main() {
	// Create registry and register substitutions
	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.TitleSubstituter{})
	registry.Register(&substitution.ContentSubstituter{})

	// Generate site with validators
	gen := generator.New(registry).
		WithValidator(validator.NewImageValidator())

	if err := gen.Generate("content", "target/build"); err != nil {
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}
}
