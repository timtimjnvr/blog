package site

import (
	"fmt"
	"os"
	"path/filepath"
)

// copyDir copies files from srcDir to destDir, optionally filtering by the provided function.
// If filter is nil, all files are copied. If filter returns true, the file is copied.
func copyDir(srcDir, destDir string, filter func(path string) bool) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filter != nil && !filter(path) {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, relPath)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
		fmt.Printf("Copied: %s -> %s\n", relPath, outPath)
		return nil
	})
}
