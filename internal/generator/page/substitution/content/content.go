package content

import (
	"github.com/timtimjnvr/blog/internal/markdown"
)

// Substituter resolves {{content}} placeholder
// it replaces links and assets with their real path in the build directory
type Substituter struct {
}

func NewSubstituer() Substituter {
	return Substituter{}

}

func (c Substituter) Placeholder() string {
	return "{{content}}"
}

func (c Substituter) Resolve(htmlContent string) (string, error) {
	htmlContent = markdown.ConvertMdLinksToHtml(htmlContent, "")
	htmlContent = markdown.ConvertAssetPaths(htmlContent, "")
	return htmlContent, nil
}
