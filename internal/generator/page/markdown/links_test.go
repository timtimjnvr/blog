package markdown

import "testing"

func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "index.md becomes index.html",
			input:    "index.md",
			expected: "index.html",
		},
		{
			name:     "posts/index.md becomes posts/index.html",
			input:    "posts/index.md",
			expected: "posts/index.html",
		},
		{
			name:     "posts/hello.md becomes posts/hello.html",
			input:    "posts/hello.md",
			expected: "posts/hello.html",
		},
		{
			name:     "other.md becomes other.html",
			input:    "other.md",
			expected: "other.html",
		},
		{
			name:     "about/index.md becomes about/index.html",
			input:    "about/index.md",
			expected: "about/index.html",
		},
		{
			name:     "nested/page.md becomes nested/page.html",
			input:    "nested/page.md",
			expected: "nested/page.html",
		},
		{
			name:     "deeply nested path preserved",
			input:    "docs/api/reference.md",
			expected: "docs/api/reference.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveOutputPath(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveOutputPath(%q) = %q, want %q", tt.input, result, tt.expected)
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
			name:          "index to posts/index",
			html:          `<a href="posts/index.md">Articles</a>`,
			sourceRelPath: "index.md",
			expected:      `<a href="posts/index.html">Articles</a>`,
		},
		{
			name:          "posts/index to sibling post",
			html:          `<a href="hello.md">Hello</a>`,
			sourceRelPath: "posts/index.md",
			expected:      `<a href="hello.html">Hello</a>`,
		},
		{
			name:          "post to posts/index",
			html:          `<a href="index.md">Back</a>`,
			sourceRelPath: "posts/hello.md",
			expected:      `<a href="index.html">Back</a>`,
		},
		{
			name:          "post to root index (parent directory)",
			html:          `<a href="../index.md">Accueil</a>`,
			sourceRelPath: "posts/hello.md",
			expected:      `<a href="../index.html">Accueil</a>`,
		},
		{
			name:          "external links unchanged",
			html:          `<a href="https://example.com">External</a>`,
			sourceRelPath: "index.md",
			expected:      `<a href="https://example.com">External</a>`,
		},
		{
			name:          "non-md links unchanged",
			html:          `<a href="image.png">Image</a>`,
			sourceRelPath: "index.md",
			expected:      `<a href="image.png">Image</a>`,
		},
		{
			name:          "multiple links in same html",
			html:          `<a href="posts/index.md">Posts</a> and <a href="about/index.md">About</a>`,
			sourceRelPath: "index.md",
			expected:      `<a href="posts/index.html">Posts</a> and <a href="about/index.html">About</a>`,
		},
		{
			name:          "link with other attributes",
			html:          `<a class="nav" href="posts/index.md" title="Posts">Articles</a>`,
			sourceRelPath: "index.md",
			expected:      `<a class="nav" href="posts/index.html" title="Posts">Articles</a>`,
		},
		{
			name:          "no links in html",
			html:          `<p>Just text</p>`,
			sourceRelPath: "index.md",
			expected:      `<p>Just text</p>`,
		},
		{
			name:          "empty html",
			html:          ``,
			sourceRelPath: "index.md",
			expected:      ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMdLinksToHtml(tt.html, tt.sourceRelPath)
			if result != tt.expected {
				t.Errorf("ConvertMdLinksToHtml(%q, %q) = %q, want %q",
					tt.html, tt.sourceRelPath, result, tt.expected)
			}
		})
	}
}

func TestConvertMdLinksToHtml_EmptySourceRelPath(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple md link replacement",
			html:     `<a href="page.md">Page</a>`,
			expected: `<a href="page.html">Page</a>`,
		},
		{
			name:     "nested md link replacement",
			html:     `<a href="posts/hello.md">Hello</a>`,
			expected: `<a href="posts/hello.html">Hello</a>`,
		},
		{
			name:     "parent path md link replacement",
			html:     `<a href="../index.md">Home</a>`,
			expected: `<a href="../index.html">Home</a>`,
		},
		{
			name:     "multiple md links",
			html:     `<a href="a.md">A</a> <a href="b.md">B</a>`,
			expected: `<a href="a.html">A</a> <a href="b.html">B</a>`,
		},
		{
			name:     "no md links unchanged",
			html:     `<a href="https://example.com">Ext</a>`,
			expected: `<a href="https://example.com">Ext</a>`,
		},
		{
			name:     "empty html",
			html:     ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMdLinksToHtml(tt.html, "")
			if result != tt.expected {
				t.Errorf("ConvertMdLinksToHtml(%q, \"\") = %q, want %q",
					tt.html, result, tt.expected)
			}
		})
	}
}

func TestConvertAssetPaths(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		sourceRelPath string
		expected      string
	}{
		{
			name:          "absolute path unchanged",
			html:          `<img src="/images/logo.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="/images/logo.png">`,
		},
		{
			name:          "external url unchanged",
			html:          `<img src="https://example.com/img.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="https://example.com/img.png">`,
		},
		{
			name:          "http url unchanged",
			html:          `<img src="http://example.com/img.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="http://example.com/img.png">`,
		},
		{
			name:          "relative path without parent ref unchanged",
			html:          `<img src="image.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="image.png">`,
		},
		{
			name:          "parent relative path stays relative to output dir",
			html:          `<img src="../assets/image.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="../assets/image.png">`,
		},
		{
			name:          "no img tags unchanged",
			html:          `<p>No images</p>`,
			sourceRelPath: "index.md",
			expected:      `<p>No images</p>`,
		},
		{
			name:          "empty html",
			html:          ``,
			sourceRelPath: "index.md",
			expected:      ``,
		},
		{
			name:          "multiple img tags with parent refs",
			html:          `<img src="../assets/a.png"> and <img src="../assets/b.png">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img src="../assets/a.png"> and <img src="../assets/b.png">`,
		},
		{
			name:          "img with other attributes and parent ref",
			html:          `<img class="w-full" src="../assets/photo.jpg" alt="Photo">`,
			sourceRelPath: "posts/hello.md",
			expected:      `<img class="w-full" src="../assets/photo.jpg" alt="Photo">`,
		},
		{
			name:          "deeply nested source with parent ref adjusts path",
			html:          `<img src="../../assets/image.png">`,
			sourceRelPath: "content/posts/hello.md",
			expected:      `<img src="../../assets/image.png">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertAssetPaths(tt.html, tt.sourceRelPath)
			if result != tt.expected {
				t.Errorf("ConvertAssetPaths(%q, %q) = %q, want %q",
					tt.html, tt.sourceRelPath, result, tt.expected)
			}
		})
	}
}

func TestConvertMdLinksToHtml_RelativePathCalculation(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		sourceRelPath string
		expected      string
	}{
		{
			name:          "from root to nested",
			html:          `<a href="posts/hello.md">Hello</a>`,
			sourceRelPath: "index.md",
			expected:      `<a href="posts/hello.html">Hello</a>`,
		},
		{
			name:          "from posts/index to sibling post",
			html:          `<a href="second-post.md">Second</a>`,
			sourceRelPath: "posts/index.md",
			expected:      `<a href="second-post.html">Second</a>`,
		},
		{
			name:          "from post back to posts index",
			html:          `<a href="index.md">Index</a>`,
			sourceRelPath: "posts/hello.md",
			expected:      `<a href="index.html">Index</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMdLinksToHtml(tt.html, tt.sourceRelPath)
			if result != tt.expected {
				t.Errorf("ConvertMdLinksToHtml(%q, %q) = %q, want %q",
					tt.html, tt.sourceRelPath, result, tt.expected)
			}
		})
	}
}
