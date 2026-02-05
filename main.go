package main

import (
	"log"

	"github.com/timtimjnvr/blog/internal/generator/site"
)

func main() {
	gen := site.NewGenerator()

	if err := gen.Generate(); err != nil {
		log.Fatalf("Generation Error: %v\n", err)
	}

	if err := gen.Validate(); err != nil {
		log.Fatalf("Validation Error: %v\n", err)
	}

	log.Println("Site generated successfully")
}
