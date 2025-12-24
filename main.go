package main

import (
	_ "embed"
	"io"
	"os"

	"github.com/timtimjnvr/blog/src/page"
)

//go:embed data/index.html
var indexHTML []byte

func main() {
	sourceMd, err := os.Open("content/index.md")
	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(sourceMd)
	if err != nil {
		panic(err)
	}

	p, err := page.Parse(b)
	if err != nil {
		panic(err)
	}

	data, err := page.Substitute(indexHTML, p)
	if err != nil {
		panic(err)
	}

	destinationFile, err := os.Create("build/index.html")
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = destinationFile.Close()
	}()

	_, err = destinationFile.Write(data)
	if err != nil {
		panic(err)
	}

	err = sourceMd.Close()
	if err != nil {
		panic(err)
	}
}
