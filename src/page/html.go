package page

import (
	"bytes"
	"fmt"
)

const content = "content"

func Substitute(html []byte, p Page) ([]byte, error) {
	if len(html) == 0 {
		return nil, fmt.Errorf("empty html")
	}
	html = bytes.Replace(html, []byte(fmt.Sprintf(`{{%s}}`, title)), []byte(p.Title), 1)
	html = bytes.Replace(html, []byte(fmt.Sprintf("{{%s}}", content)), p.Content, 1)
	return html, nil
}
