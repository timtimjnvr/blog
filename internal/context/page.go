package context

// Context defines what data substitutions can access
type Context interface {
	GetSource() []byte
	GetRelPath() string
	GetHTMLContent() string
}

// PageContext contains all data available for substitutions
type PageContext struct {
	Source      []byte // raw markdown content
	RelPath     string // relative path of source file
	HTMLContent string // HTML content after goldmark conversion
}

func (p *PageContext) GetSource() []byte {
	return p.Source
}

func (p *PageContext) GetRelPath() string {
	return p.RelPath
}

func (p *PageContext) GetHTMLContent() string {
	return p.HTMLContent
}
