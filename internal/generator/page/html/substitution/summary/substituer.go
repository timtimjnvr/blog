package summary

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var headingRe = regexp.MustCompile(`<h([2-6])[^>]*id="([^"]+)"[^>]*>([^<]+)(?:<a[^>]*>[^<]*</a>)?</h[2-6]>`)

func textSizeClass(depth int) string {
	switch depth {
	case 1:
		return "text-base"
	case 2:
		return "text-sm"
	default:
		return "text-xs"
	}
}

// Substituter resolves the {{summary}} placeholder with a generated table of contents.
type Substituter struct{}

func NewSubstituer() Substituter {
	return Substituter{}
}

func (s Substituter) Placeholder() string {
	return "<p>{{summary}}</p>"
}

func (s Substituter) Resolve(content string) (string, error) {
	matches := headingRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return "", nil
	}

	type heading struct {
		level int
		id    string
		text  string
	}

	items := make([]heading, len(matches))
	for i, m := range matches {
		level, _ := strconv.Atoi(m[1])
		items[i] = heading{level: level, id: m[2], text: strings.TrimSpace(m[3])}
	}

	var sb strings.Builder
	sb.WriteString("<nav>")

	prevLevel := items[0].level - 1
	depth := 0

	for _, item := range items {
		switch {
		case item.level > prevLevel:
			for i := prevLevel; i < item.level; i++ {
				sb.WriteString(`<ul class="space-y-1">`)
				depth++
			}
		case item.level == prevLevel:
			sb.WriteString("</li>")
		default:
			sb.WriteString("</li>")
			for i := item.level; i < prevLevel; i++ {
				sb.WriteString("</ul></li>")
				depth--
			}
		}
		fmt.Fprintf(&sb, `<li><a href="#%s" class="%s">%s</a>`, item.id, textSizeClass(depth), item.text)
		prevLevel = item.level
	}

	sb.WriteString("</li>")
	for depth > 1 {
		sb.WriteString("</ul></li>")
		depth--
	}
	sb.WriteString("</ul></nav>")

	return sb.String(), nil
}
