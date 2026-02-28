package main

import (
	"log"

	"github.com/timtimjnvr/blog/internal/generator/site"
)

func main() {
	gen, err := site.NewGenerator()
	if err != nil {
		log.Fatalf("Could not create the site generator: %v\n", err)
	}

	if err := gen.Generate(); err != nil {
		log.Fatalf("Site generation error: %v\n", err)
	}

	if err := gen.Validate(); err != nil {
		log.Fatalf("Site validation error: %v\n", err)
	}

	log.Println("Site generated successfully !")
}
