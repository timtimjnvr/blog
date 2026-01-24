package main

import (
	"os"
	"strings"
	"testing"
)

func TestSiteGeneration(t *testing.T) {
	// Clean build directory first
	os.RemoveAll("build")
	os.MkdirAll("build", 0755)

	// Run the main function
	main()

	// Check that expected files were generated
	tests := []struct {
		path     string
		contains []string
	}{
		{
			path: "build/index.html",
			contains: []string{
				"<h1>Bienvenue sur mon blog</h1>",
				`<a href="posts/index.html">Articles</a>`,
				"<!DOCTYPE html>",
			},
		},
		{
			path: "build/posts/index.html",
			contains: []string{
				"<h1>Articles</h1>",
				`<a href="../post/hello.html">Hello World</a>`,
				`<a href="../post/second-post.html">`,
			},
		},
		{
			path: "build/post/hello.html",
			contains: []string{
				"<h1>Hello World</h1>",
				`<a href="../posts/index.html">Articles</a>`,
				"<pre><code class=\"language-go\">",
			},
		},
		{
			path: "build/post/second-post.html",
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
	os.RemoveAll("build")
	os.MkdirAll("build", 0755)
	main()

	// Define expected links and their targets
	linkTests := []struct {
		sourcePath string
		linkHref   string
		targetPath string
	}{
		{"build/index.html", `href="posts/index.html"`, "build/posts/index.html"},
		{"build/posts/index.html", `href="../post/hello.html"`, "build/post/hello.html"},
		{"build/posts/index.html", `href="../index.html"`, "build/index.html"},
		{"build/post/hello.html", `href="../posts/index.html"`, "build/posts/index.html"},
		{"build/post/second-post.html", `href="../index.html"`, "build/index.html"},
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

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"# Hello World\n\nContent here", "Hello World"},
		{"Some text\n# Title Here\n\nMore", "Title Here"},
		{"No heading here", "Untitled"},
		{"## Only H2\n\nNo H1", "Untitled"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := extractTitle([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("extractTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertMdLinksToHtml(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		sourceRelPath string
		expected      string
	}{
		{
			name:          "home to posts/index",
			html:          `<a href="posts/index.md">Articles</a>`,
			sourceRelPath: "home.md",
			expected:      `<a href="posts/index.html">Articles</a>`,
		},
		{
			name:          "posts/index to post",
			html:          `<a href="hello.md">Hello</a>`,
			sourceRelPath: "posts/index.md",
			expected:      `<a href="../post/hello.html">Hello</a>`,
		},
		{
			name:          "post to posts/index",
			html:          `<a href="index.md">Back</a>`,
			sourceRelPath: "posts/hello.md",
			expected:      `<a href="../posts/index.html">Back</a>`,
		},
		{
			name:          "post to home",
			html:          `<a href="../home.md">Accueil</a>`,
			sourceRelPath: "posts/hello.md",
			expected:      `<a href="../index.html">Accueil</a>`,
		},
		{
			name:          "external links unchanged",
			html:          `<a href="https://example.com">External</a>`,
			sourceRelPath: "home.md",
			expected:      `<a href="https://example.com">External</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMdLinksToHtml(tt.html, tt.sourceRelPath)
			if result != tt.expected {
				t.Errorf("convertMdLinksToHtml(%q, %q) = %q, want %q",
					tt.html, tt.sourceRelPath, result, tt.expected)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"home.md", "index.html"},
		{"posts/index.md", "posts/index.html"},
		{"posts/hello.md", "post/hello.html"},
		{"posts/my-article.md", "post/my-article.html"},
		{"other.md", "other.html"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := resolveOutputPath(tt.input)
			if result != tt.expected {
				t.Errorf("resolveOutputPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestApplySubstitutions(t *testing.T) {
	template := `<title>{{title}}</title><body>{{content}}</body>`

	ctx := &PageContext{
		Source:      []byte("# Test Title\n\nSome content"),
		RelPath:     "home.md",
		HTMLContent: `<h1>Test Title</h1><p>Some content</p>`,
	}

	result := applySubstitutions(template, ctx)

	if !strings.Contains(result, "<title>Test Title</title>") {
		t.Errorf("expected title substitution, got %q", result)
	}

	if !strings.Contains(result, "<body><h1>Test Title</h1><p>Some content</p></body>") {
		t.Errorf("expected content substitution, got %q", result)
	}
}

func TestApplySubstitutionsWithLinks(t *testing.T) {
	template := `{{content}}`

	ctx := &PageContext{
		Source:      []byte("# Test\n\n[Link](posts/index.md)"),
		RelPath:     "home.md",
		HTMLContent: `<a href="posts/index.md">Link</a>`,
	}

	result := applySubstitutions(template, ctx)

	expected := `<a href="posts/index.html">Link</a>`
	if result != expected {
		t.Errorf("applySubstitutions with links: got %q, want %q", result, expected)
	}
}
