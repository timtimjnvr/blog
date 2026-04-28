package markdown

import (
	"strings"
	"testing"
)

func TestD2Renderer(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	t.Run("valid d2 block produces div and svg", func(t *testing.T) {
		result, err := converter.Convert([]byte("```d2\nA -> B\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, "display: flex") {
			t.Errorf("expected flex div, got: %s", result)
		}
		if !strings.Contains(result, "<svg") {
			t.Errorf("expected SVG in output, got: %s", result)
		}
	})

	t.Run("invalid d2 block returns error", func(t *testing.T) {
		_, err := converter.Convert([]byte("```d2\n{\n```"))
		if err == nil {
			t.Fatal("expected error for invalid D2 source, got nil")
		}
	})

	t.Run("width attr adds max-width to div style", func(t *testing.T) {
		result, err := converter.Convert([]byte("```d2 width=600\nA -> B\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, "max-width: 600px") {
			t.Errorf("expected max-width: 600px in style, got: %s", result)
		}
	})

	t.Run("scale attr produces explicit SVG dimensions", func(t *testing.T) {
		result, err := converter.Convert([]byte("```d2 scale=0.5\nA -> B\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if _, _, ok := extractExplicitDims(result); !ok {
			t.Errorf("expected explicit SVG dimensions when scale=0.5, got: %s", result)
		}
	})

	t.Run("no attrs produces no max-width", func(t *testing.T) {
		result, err := converter.Convert([]byte("```d2\nA -> B\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if strings.Contains(result, "max-width") {
			t.Errorf("unexpected max-width without width attr, got: %s", result)
		}
	})
}

func TestRenderCodeBlock(t *testing.T) {
	converter := mustNewConverter(t, nil, "")

	t.Run("with language produces class attribute", func(t *testing.T) {
		result, err := converter.Convert([]byte("```go\nfunc main() {}\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, `<pre><code class="language-go">`) {
			t.Errorf("expected language class, got: %s", result)
		}
	})

	t.Run("without language produces no class attribute", func(t *testing.T) {
		result, err := converter.Convert([]byte("```\nfoo\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if !strings.Contains(result, "<pre><code>") {
			t.Errorf("expected <pre><code> without class, got: %s", result)
		}
		if strings.Contains(result, "class=") {
			t.Errorf("unexpected class attribute on code block without language, got: %s", result)
		}
	})

	t.Run("HTML content in code block is escaped", func(t *testing.T) {
		result, err := converter.Convert([]byte("```go\n<script>alert('xss')</script>\n```"))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}
		if strings.Contains(result, "<script>") {
			t.Errorf("raw <script> tag should be escaped, got: %s", result)
		}
		if !strings.Contains(result, "&lt;script&gt;") {
			t.Errorf("expected &lt;script&gt; in output, got: %s", result)
		}
	})
}
