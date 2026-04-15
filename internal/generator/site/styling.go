package site

import (
	"fmt"

	"github.com/timtimjnvr/blog/internal/generator/page/styling"
)

func (g *Generator) loadStylingConfig() error {
	if _, err := g.fs.Stat(g.optionalStylingConfigPath); err == nil {
		styleConfig, err := styling.LoadConfig(g.optionalStylingConfigPath)
		if err != nil {
			return fmt.Errorf("failed to LoadConfig: %w", err)
		}
		fmt.Printf("Loaded style configuration from %s\n", g.optionalStylingConfigPath)
		g.stylingConfig = styleConfig
	}
	return nil
}
