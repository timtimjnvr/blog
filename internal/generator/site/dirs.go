package site

import (
	"fmt"
	"path/filepath"
)

func (g *Generator) makeAllDirectories() error {
	for _, d := range []string{g.assetsOutDir, g.scriptsOutDir, g.buildDir} {
		if err := g.fs.MkdirAll(filepath.Dir(d), 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	return nil
}
