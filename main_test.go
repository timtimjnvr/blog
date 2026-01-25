package main

import (
	"os"
	"strings"
	"testing"
)

// Integration tests that verify the full site generation pipeline

func TestSiteGeneration(t *testing.T) {
	// Clean build directory first
	_ = os.RemoveAll("target/build")
	_ = os.MkdirAll("target/build", 0755)

	// Run the main function
	main()

	// Check that expected files were generated
	tests := []struct {
		path     string
		contains []string
	}{
		{
			path: "target/build/index.html",
			contains: []string{
				"<h1>Bienvenue sur mon blog</h1>",
				`<a href="posts/index.html">Articles</a>`,
				"<!DOCTYPE html>",
			},
		},
		{
			path: "target/build/posts/index.html",
			contains: []string{
				"<h1>Articles</h1>",
				`<a href="../post/hello.html">Hello World</a>`,
				`<a href="../post/second-post.html">`,
			},
		},
		{
			path: "target/build/post/hello.html",
			contains: []string{
				"<h1>Hello World</h1>",
				`<a href="../posts/index.html">Articles</a>`,
				"<pre><code class=\"language-go\">",
			},
		},
		{
			path: "target/build/post/second-post.html",
			contains: []string{
				"<h1>Deuxième article</h1>",
				`<a href="../index.html">Accueil</a>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			content, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.path, err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(string(content), substr) {
					t.Errorf("%s should contain %q", tt.path, substr)
				}
			}
		})
	}
}

func TestMarkdownFilesExist(t *testing.T) {
	mdFiles := []string{
		"content/home.md",
		"content/posts/index.md",
		"content/posts/hello.md",
		"content/posts/second-post.md",
	}

	for _, path := range mdFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("markdown file %s does not exist", path)
		}
	}
}

func TestMarkdownFilesContainNoHtmlLinks(t *testing.T) {
	mdFiles := []string{
		"content/home.md",
		"content/posts/index.md",
		"content/posts/hello.md",
		"content/posts/second-post.md",
	}

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}

		if strings.Contains(string(content), ".html") {
			t.Errorf("%s should not contain .html links, only .md links", path)
		}
	}
}

func TestLinksAreValid(t *testing.T) {
	// Clean and regenerate
	_ = os.RemoveAll("target/build")
	_ = os.MkdirAll("target/build", 0755)
	main()

	// Define expected links and their targets
	linkTests := []struct {
		sourcePath string
		linkHref   string
		targetPath string
	}{
		{"target/build/index.html", `href="posts/index.html"`, "target/build/posts/index.html"},
		{"target/build/posts/index.html", `href="../post/hello.html"`, "target/build/post/hello.html"},
		{"target/build/posts/index.html", `href="../index.html"`, "target/build/index.html"},
		{"target/build/post/hello.html", `href="../posts/index.html"`, "target/build/posts/index.html"},
		{"target/build/post/second-post.html", `href="../index.html"`, "target/build/index.html"},
	}

	for _, tt := range linkTests {
		t.Run(tt.sourcePath+"->"+tt.targetPath, func(t *testing.T) {
			content, err := os.ReadFile(tt.sourcePath)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.sourcePath, err)
			}

			if strings.Contains(string(content), tt.linkHref) {
				if _, err := os.Stat(tt.targetPath); os.IsNotExist(err) {
					t.Errorf("%s links to %s but target does not exist", tt.sourcePath, tt.targetPath)
				}
			}
		})
	}
}
