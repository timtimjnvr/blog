package markdown

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// testCompiler is shared across tests to avoid re-initialising the ruler
// (which loads font data) for every test case.
var (
	testCompiler     *d2Compiler
	testCompilerOnce sync.Once
	testCompilerErr  error
)

func getTestCompiler(t *testing.T) *d2Compiler {
	t.Helper()
	testCompilerOnce.Do(func() {
		testCompiler, testCompilerErr = newD2Compiler()
	})
	if testCompilerErr != nil {
		t.Fatalf("newD2Compiler: %v", testCompilerErr)
	}
	return testCompiler
}

// --- helpers ----------------------------------------------------------------

var (
	reOuterSVG     = regexp.MustCompile(`<svg[^>]*data-d2-version[^>]*>`)
	reViewBox      = regexp.MustCompile(`viewBox="0 0 (\d+) (\d+)"`)
	reExplicitDims = regexp.MustCompile(`\bwidth="(\d+)"\s+height="(\d+)"`)
)

func outerSVGTag(svg string) string {
	return reOuterSVG.FindString(svg)
}

func extractViewBox(svg string) (w, h int, ok bool) {
	m := reViewBox.FindStringSubmatch(outerSVGTag(svg))
	if len(m) < 3 {
		return 0, 0, false
	}
	w, _ = strconv.Atoi(m[1])
	h, _ = strconv.Atoi(m[2])
	return w, h, true
}

func extractExplicitDims(svg string) (w, h int, ok bool) {
	m := reExplicitDims.FindStringSubmatch(outerSVGTag(svg))
	if len(m) < 3 {
		return 0, 0, false
	}
	w, _ = strconv.Atoi(m[1])
	h, _ = strconv.Atoi(m[2])
	return w, h, true
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// --- tests ------------------------------------------------------------------

func TestNewD2Compiler(t *testing.T) {
	c, err := newD2Compiler()
	if err != nil {
		t.Fatalf("newD2Compiler() error = %v", err)
	}
	if c.ruler == nil {
		t.Error("ruler is nil")
	}
	if c.ctx == nil {
		t.Error("ctx is nil")
	}
}

func TestD2Compiler_Compile(t *testing.T) {
	c := getTestCompiler(t)

	t.Run("valid source produces SVG", func(t *testing.T) {
		svg, err := c.compile("A -> B", 0)
		if err != nil {
			t.Fatalf("compile() error = %v", err)
		}
		if !strings.Contains(svg, "<svg") {
			t.Errorf("expected SVG in output, got: %s", svg)
		}
	})

	t.Run("invalid source returns wrapped error", func(t *testing.T) {
		// Unclosed brace is a reliable D2 parse error.
		_, err := c.compile("{", 0)
		if err == nil {
			t.Fatal("expected error for invalid D2 source")
		}
		if !strings.Contains(err.Error(), "d2 compile:") {
			t.Errorf("expected error prefixed with \"d2 compile:\", got: %v", err)
		}
	})

	t.Run("empty source documents behavior", func(t *testing.T) {
		// D2 produces a minimal valid SVG for empty input rather than an error.
		svg, err := c.compile("", 0)
		if err != nil {
			t.Logf("empty source returns error (documenting actual behavior): %v", err)
			return
		}
		if !strings.Contains(svg, "<svg") {
			t.Errorf("expected SVG for empty source, got: %s", svg)
		}
	})
}

func TestD2Compiler_Compile_Scale(t *testing.T) {
	c := getTestCompiler(t)
	const src = "A -> B"

	t.Run("zero scale produces no explicit dimensions", func(t *testing.T) {
		svg, err := c.compile(src, 0)
		if err != nil {
			t.Fatalf("compile() error = %v", err)
		}
		if tag := outerSVGTag(svg); strings.Contains(tag, `width="`) {
			t.Errorf("expected no explicit dimensions when scale=0, outer tag: %s", tag)
		}
	})

	t.Run("positive scale produces explicit dimensions", func(t *testing.T) {
		svg, err := c.compile(src, 0.5)
		if err != nil {
			t.Fatalf("compile() error = %v", err)
		}
		if tag := outerSVGTag(svg); !strings.Contains(tag, `width="`) {
			t.Errorf("expected explicit dimensions when scale=0.5, outer tag: %s", tag)
		}
	})

	t.Run("scale 0.5 produces dimensions half of scale 1.0", func(t *testing.T) {
		svgFull, err := c.compile(src, 1.0)
		if err != nil {
			t.Fatalf("compile(scale=1.0) error = %v", err)
		}
		svgHalf, err := c.compile(src, 0.5)
		if err != nil {
			t.Fatalf("compile(scale=0.5) error = %v", err)
		}

		wFull, hFull, ok := extractExplicitDims(svgFull)
		if !ok {
			t.Fatal("could not extract dimensions from scale=1.0 SVG")
		}
		wHalf, hHalf, ok := extractExplicitDims(svgHalf)
		if !ok {
			t.Fatal("could not extract dimensions from scale=0.5 SVG")
		}

		// D2 uses math.Ceil, so allow ±1 pixel tolerance.
		if absInt(wHalf*2-wFull) > 1 {
			t.Errorf("width: scale=0.5 gave %d, expected ~%d (half of %d)", wHalf, wFull/2, wFull)
		}
		if absInt(hHalf*2-hFull) > 1 {
			t.Errorf("height: scale=0.5 gave %d, expected ~%d (half of %d)", hHalf, hFull/2, hFull)
		}
	})
}

func TestD2Compiler_Compile_Direction(t *testing.T) {
	c := getTestCompiler(t)
	// A long chain ensures the aspect ratio difference between
	// vertical and horizontal layout is unambiguous.
	const chain = "A -> B -> C -> D -> E"

	t.Run("default layout is vertical", func(t *testing.T) {
		svg, err := c.compile(chain, 0)
		if err != nil {
			t.Fatalf("compile() error = %v", err)
		}
		w, h, ok := extractViewBox(svg)
		if !ok {
			t.Fatal("could not extract viewBox")
		}
		if w >= h {
			t.Errorf("expected vertical layout (height > width), got viewBox %dx%d", w, h)
		}
	})

	t.Run("direction right produces horizontal layout", func(t *testing.T) {
		svg, err := c.compile("direction: right\n"+chain, 0)
		if err != nil {
			t.Fatalf("compile() error = %v", err)
		}
		w, h, ok := extractViewBox(svg)
		if !ok {
			t.Fatal("could not extract viewBox")
		}
		if w <= h {
			t.Errorf("expected horizontal layout (width > height), got viewBox %dx%d", w, h)
		}
	})
}
