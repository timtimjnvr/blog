package metadata

import (
	"regexp"
)

type Metadata struct {
	CreationDate string
}

var metadataRegexp = regexp.MustCompile(`<!--\s*(\S+):\s*(.+?)\s*-->`)

func Extract(data []byte) Metadata {
	var m Metadata
	for _, match := range metadataRegexp.FindAllSubmatch(data, -1) {
		key := string(match[1])
		value := string(match[2])
		switch key {
		case "creation-date":
			m.CreationDate = value
		}
	}
	return m
}
