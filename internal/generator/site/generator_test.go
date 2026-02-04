package site

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// mockPageGenerator is a test double for PageGenerator
type mockPageGenerator struct {
	err error
}

func (m *mockPageGenerator) Generate() error {
	return m.err
}

// createDirs creates the base directory and any subdirectories relative to it
func createDirs(t *testing.T, base string, dirs ...string) {
	t.Helper()
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(base, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
}

// createFiles creates files relative to the base path
func createFiles(t *testing.T, base string, files ...string) {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(base, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerator_listSections(t *testing.T) {
	tests := []struct {
		name             string
		dirs             []string
		files            []string
		createContentDir bool
		expectedSections []string // relative to content dir ("." for root)
		wantErr          bool
	}{
		{
			name:             "empty content directory",
			createContentDir: true,
			expectedSections: []string{"."},
		},
		{
			name:             "single top-level section",
			dirs:             []string{"blog"},
			expectedSections: []string{".", "blog"},
		},
		{
			name:             "multiple top-level sections",
			dirs:             []string{"blog", "about", "projects"},
			expectedSections: []string{".", "about", "blog", "projects"},
		},
		{
			name:             "nested directories are excluded",
			dirs:             []string{"blog", "blog/2024", "blog/2024/january"},
			expectedSections: []string{".", "blog"},
		},
		{
			name:             "files are ignored",
			dirs:             []string{"blog"},
			files:            []string{"index.md", "blog/post.md"},
			expectedSections: []string{".", "blog"},
		},
		{
			name:    "non-existent content directory",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			contentDir := filepath.Join(tmpDir, "content")

			if tt.createContentDir || len(tt.dirs) > 0 || len(tt.files) > 0 {
				createDirs(t, contentDir, tt.dirs...)
			}
			createFiles(t, contentDir, tt.files...)

			g := &Generator{
				ContentDir:            contentDir,
				SectionDirectoryNames: make([]string, 0),
			}

			err := g.listSections()

			if (err != nil) != tt.wantErr {
				t.Errorf("listSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Convert expected sections to full paths
			expectedPaths := make(map[string]bool)
			for _, section := range tt.expectedSections {
				if section == "." {
					expectedPaths[contentDir] = true
				} else {
					expectedPaths[filepath.Join(contentDir, section)] = true
				}
			}

			if len(g.SectionDirectoryNames) != len(expectedPaths) {
				t.Errorf("got %d sections, want %d\ngot: %v", len(g.SectionDirectoryNames), len(expectedPaths), g.SectionDirectoryNames)
				return
			}

			for _, path := range g.SectionDirectoryNames {
				if !expectedPaths[path] {
					t.Errorf("unexpected section %q", path)
				}
			}
		})
	}
}

// listFiles returns all file paths relative to base directory
func listFiles(t *testing.T, base string) []string {
	t.Helper()
	var files []string
	filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(base, path)
		files = append(files, relPath)
		return nil
	})
	slices.Sort(files)
	return files
}

func TestGenerator_copyAssets(t *testing.T) {
	tests := []struct {
		name          string
		srcFiles      []string
		expectedFiles []string
		wantErr       bool
	}{
		{
			name:          "copies all files",
			srcFiles:      []string{"image.png", "style.css", "data.json"},
			expectedFiles: []string{"data.json", "image.png", "style.css"},
		},
		{
			name:          "preserves directory structure",
			srcFiles:      []string{"images/logo.png", "images/icons/favicon.ico", "docs/guide.pdf"},
			expectedFiles: []string{"docs/guide.pdf", "images/icons/favicon.ico", "images/logo.png"},
		},
		{
			name:          "empty assets directory",
			srcFiles:      []string{},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			assetsDir := filepath.Join(tmpDir, "assets")
			buildDir := filepath.Join(tmpDir, "build")

			createDirs(t, assetsDir)
			createFiles(t, assetsDir, tt.srcFiles...)

			g := &Generator{
				AssetsDir:    assetsDir,
				AssetsOutDir: "assets",
				BuildDir:     buildDir,
			}

			err := g.copyAssets()

			if (err != nil) != tt.wantErr {
				t.Errorf("copyAssets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			destDir := filepath.Join(buildDir, g.AssetsOutDir)
			gotFiles := listFiles(t, destDir)

			if len(gotFiles) != len(tt.expectedFiles) {
				t.Errorf("got %d files, want %d\ngot: %v\nwant: %v",
					len(gotFiles), len(tt.expectedFiles), gotFiles, tt.expectedFiles)
				return
			}

			for i, got := range gotFiles {
				if got != tt.expectedFiles[i] {
					t.Errorf("file mismatch at index %d: got %q, want %q", i, got, tt.expectedFiles[i])
				}
			}
		})
	}
}

func TestGenerator_generatePages_WithWrongExtension_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	buildDir := filepath.Join(tmpDir, "build")

	createDirs(t, contentDir)
	createFiles(t, contentDir, "page.txt")

	g := &Generator{
		ContentDir: contentDir,
		BuildDir:   buildDir,
		pageGeneratorFactory: func(markdownPath, buildDir string) PageGenerator {
			return &mockPageGenerator{}
		},
	}

	err := g.generatePages()

	if err == nil {
		t.Fatal("expected error for wrong extension, got nil")
	}

	if !errors.Is(err, err) || err.Error() == "" {
		t.Errorf("expected error message about wrong extension, got: %v", err)
	}
}

func TestGenerator_generatePages_WithPageGeneratorFailure_ReturnsErrors(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	buildDir := filepath.Join(tmpDir, "build")

	createDirs(t, contentDir)
	createFiles(t, contentDir, "page1.md", "page2.md")

	expectedErr := errors.New("page generation failed")

	g := &Generator{
		ContentDir: contentDir,
		BuildDir:   buildDir,
		pageGeneratorFactory: func(markdownPath, buildDir string) PageGenerator {
			return &mockPageGenerator{err: expectedErr}
		},
	}

	err := g.generatePages()

	if err == nil {
		t.Fatal("expected error when page generator fails, got nil")
	}

	// errors.Join combines multiple errors, check that our error is in there
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to contain %q, got: %v", expectedErr, err)
	}
}

func TestGenerator_generatePages_WithValidFiles_ReturnsNilAndPopulatesGeneratedPagesPath(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	buildDir := filepath.Join(tmpDir, "build")

	createDirs(t, contentDir)
	createFiles(t, contentDir, "page1.md", "subdir/page2.md")

	g := &Generator{
		ContentDir: contentDir,
		BuildDir:   buildDir,
		pageGeneratorFactory: func(markdownPath, buildDir string) PageGenerator {
			return &mockPageGenerator{}
		},
	}

	err := g.generatePages()

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if len(g.GeneratedPagesPath) != 2 {
		t.Errorf("expected 2 pages in GeneratedPagesPath, got %d", len(g.GeneratedPagesPath))
	}
}

func TestGenerator_copyScripts(t *testing.T) {
	tests := []struct {
		name          string
		srcFiles      []string
		expectedFiles []string
		wantErr       bool
	}{
		{
			name:          "copies only js files",
			srcFiles:      []string{"app.js", "utils.js", "style.css", "readme.txt"},
			expectedFiles: []string{"app.js", "utils.js"},
		},
		{
			name:          "preserves directory structure",
			srcFiles:      []string{"main.js", "lib/helper.js", "lib/vendor/jquery.js"},
			expectedFiles: []string{"lib/helper.js", "lib/vendor/jquery.js", "main.js"},
		},
		{
			name:          "empty scripts directory",
			srcFiles:      []string{},
			expectedFiles: []string{},
		},
		{
			name:          "no js files",
			srcFiles:      []string{"style.css", "data.json"},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptsDir := filepath.Join(tmpDir, "scripts")
			buildDir := filepath.Join(tmpDir, "build")

			createDirs(t, scriptsDir)
			createFiles(t, scriptsDir, tt.srcFiles...)

			g := &Generator{
				ScriptsDir:    scriptsDir,
				ScriptsOutDir: "scripts",
				BuildDir:      buildDir,
			}

			err := g.copyScripts()

			if (err != nil) != tt.wantErr {
				t.Errorf("copyScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			destDir := filepath.Join(buildDir, g.ScriptsOutDir)
			gotFiles := listFiles(t, destDir)

			if len(gotFiles) != len(tt.expectedFiles) {
				t.Errorf("got %d files, want %d\ngot: %v\nwant: %v",
					len(gotFiles), len(tt.expectedFiles), gotFiles, tt.expectedFiles)
				return
			}

			for i, got := range gotFiles {
				if got != tt.expectedFiles[i] {
					t.Errorf("file mismatch at index %d: got %q, want %q", i, got, tt.expectedFiles[i])
				}
			}
		})
	}
}
