package markdown

import (
	"fmt"
	"strings"
	"testing"
)

func TestHeadingRenderer_AllLevels(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	for level := 1; level <= 6; level++ {
		t.Run(fmt.Sprintf("h%d", level), func(t *testing.T) {
			input := fmt.Sprintf("%s Heading", strings.Repeat("#", level))
			result, err := converter.Convert([]byte(input))
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			tag := fmt.Sprintf("<h%d", level)
			if !strings.Contains(result, tag) {
				t.Errorf("expected %s tag, got %q", tag, result)
			}

			if !strings.Contains(result, `<a href="#heading" class="heading-anchor">#</a>`) {
				t.Errorf("expected anchor link in heading, got %q", result)
			}

			closeTag := fmt.Sprintf("</h%d>", level)
			if !strings.Contains(result, closeTag) {
				t.Errorf("expected closing %s tag, got %q", closeTag, result)
			}
		})
	}
}

func TestHeadingRenderer_SlugifiedID(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	tests := []struct {
		name       string
		input      string
		expectedID string
	}{
		{
			name:       "special characters are removed",
			input:      "## Hello, World!",
			expectedID: "hello-world",
		},
		{
			name:       "spaces become hyphens",
			input:      "## My Great Section",
			expectedID: "my-great-section",
		},
		{
			name:       "mixed case becomes lowercase",
			input:      "## CamelCase Title",
			expectedID: "camelcase-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			expectedAttr := fmt.Sprintf(`id="%s"`, tt.expectedID)
			if !strings.Contains(result, expectedAttr) {
				t.Errorf("expected %s, got %q", expectedAttr, result)
			}

			expectedHref := fmt.Sprintf(`href="#%s"`, tt.expectedID)
			if !strings.Contains(result, expectedHref) {
				t.Errorf("expected anchor %s, got %q", expectedHref, result)
			}
		})
	}
}

func TestHeadingRenderer_DuplicateHeadings(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	input := "## Section\n\nSome text.\n\n## Section\n\nMore text.\n\n## Section"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `id="section"`) {
		t.Errorf("expected first heading with id=\"section\", got %q", result)
	}
	if !strings.Contains(result, `id="section-1"`) {
		t.Errorf("expected second heading with id=\"section-1\", got %q", result)
	}
	if !strings.Contains(result, `id="section-2"`) {
		t.Errorf("expected third heading with id=\"section-2\", got %q", result)
	}

	if !strings.Contains(result, `href="#section-1"`) {
		t.Errorf("expected anchor href matching duplicate id, got %q", result)
	}
}

func TestHeadingRenderer_CustomID(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	input := "## Section {#custom-id}"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `id="custom-id"`) {
		t.Errorf("expected custom id, got %q", result)
	}
	if !strings.Contains(result, `href="#custom-id"`) {
		t.Errorf("expected anchor with custom id, got %q", result)
	}
}

func TestHeadingRenderer_AnchorStructure(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	input := "## My Section"
	result, err := converter.Convert([]byte(input))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Anchor has the correct class
	if !strings.Contains(result, `class="heading-anchor"`) {
		t.Errorf("expected heading-anchor class, got %q", result)
	}

	// Anchor content is #
	if !strings.Contains(result, `">#</a>`) {
		t.Errorf("expected # as anchor content, got %q", result)
	}

	// Anchor appears after heading text
	anchorIdx := strings.Index(result, `<a href="#my-section"`)
	textIdx := strings.Index(result, "My Section")
	if anchorIdx == -1 || textIdx == -1 || anchorIdx <= textIdx {
		t.Errorf("anchor should appear after heading text, got %q", result)
	}
}
