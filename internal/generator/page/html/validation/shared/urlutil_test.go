package shared

import (
	"testing"
)

func TestIsExternalURL(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{"https://example.com/page", true},
		{"http://example.com/page", true},
		{"/assets/image.png", false},
		{"../relative/path.html", false},
		{"index.html", false},
		{"mailto:user@example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			got := IsExternalURL(tt.src)
			if got != tt.want {
				t.Errorf("IsExternalURL(%q) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestResolveLocalPath(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		htmlPath string
		buildDir string
		want     string
	}{
		{
			name:     "absolute path resolved from build root",
			src:      "/assets/image.png",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/assets/image.png",
		},
		{
			name:     "relative path resolved from html directory",
			src:      "image.png",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/posts/image.png",
		},
		{
			name:     "relative path with parent traversal",
			src:      "../assets/style.css",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/assets/style.css",
		},
		{
			name:     "absolute path at build root",
			src:      "/index.html",
			htmlPath: "/build/index.html",
			buildDir: "/build",
			want:     "/build/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLocalPath(tt.src, tt.htmlPath, tt.buildDir)
			if got != tt.want {
				t.Errorf("ResolveLocalPath(%q, %q, %q) = %q, want %q", tt.src, tt.htmlPath, tt.buildDir, got, tt.want)
			}
		})
	}
}
