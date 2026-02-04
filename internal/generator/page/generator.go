package page

type Generator struct {
	MarkdownPath string
	BuildDir     string
}

func New(markdownPath string, buildDir string) *Generator {
	return &Generator{
		MarkdownPath: markdownPath,
		BuildDir:     buildDir,
	}
}

func (g *Generator) Generate() error {
	return nil
}
