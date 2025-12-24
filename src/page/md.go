package page

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Page struct {
	Title      string
	Content    []byte
	Attributes map[string]string
}

const (
	separator = "-----------"
	title     = "title"
)

func Parse(b []byte) (Page, error) {
	if (len(b)) == 0 {
		return Page{}, errors.New("empty html")
	}

	if bytes.ContainsAny(b, separator) {
		p := Page{}
		split := bytes.Split(b, []byte(separator))
		if !bytes.ContainsAny(split[0], title) {
			return Page{}, fmt.Errorf("could not find Page title")
		}

		metadata := sanitizeBytes(split[0])
		p.Content = sanitizeBytes(split[1])

		for _, line := range bytes.Split(metadata, []byte(`\n`)) {
			kv := bytes.Split(line, []byte(`:`))
			if len(kv) != 2 {
				return Page{}, fmt.Errorf("could not Parse Page metadata")
			}

			if string(kv[0]) == title {
				p.Title = sanitizeString(string(kv[1]))
				continue
			}
			if p.Attributes == nil {
				p.Attributes = make(map[string]string)
			}
			p.Attributes[string(kv[0])] = sanitizeString(string(kv[1]))
		}

		return p, nil
	}

	return Page{}, fmt.Errorf("could not Parse Page")
}

func sanitizeString(s string) string {
	s = strings.TrimSuffix(s, `\n`)
	s = strings.TrimSpace(s)
	return s
}

func sanitizeBytes(b []byte) []byte {
	b = bytes.TrimPrefix(b, []byte(`\n`))
	b = bytes.TrimSuffix(b, []byte(`\n`))
	b = bytes.TrimSpace(b)
	return b
}
