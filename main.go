package main

import (
	"fmt"

	"github.com/timtimjnvr/blog/internal/context"
	"github.com/timtimjnvr/blog/internal/generator"
	"github.com/timtimjnvr/blog/internal/substitution"
)

func main() {
	// Create registry and register substitutions
	registry := substitution.NewRegistry[*context.PageContext]()
	registry.Register(&substitution.TitleSubstituter{})
	registry.Register(&substitution.ContentSubstituter{})

	// Generate site
	gen := generator.New(registry)
	if err := gen.Generate("content", "target/build"); err != nil {
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}
}
