package view

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed templates
var FS embed.FS

// InfoTpl is a template for rendering channel information.
var InfoTpl *template.Template

func init() {
	var err error

	InfoTpl, err = template.New("update").ParseFS(FS, "templates/channel_info.html")
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}
}

// filteredFS wraps an fs.FS to skip HTML template files
type filteredFS struct {
	inner fs.FS
}

func (f filteredFS) Open(name string) (fs.File, error) {
	// Block access to HTML template files
	if strings.HasSuffix(name, ".html") {
		return nil, fs.ErrNotExist
	}
	return f.inner.Open(name)
}

// StaticFS initializes the static file system for serving frontend files.
// HTML template files are excluded from static serving for security.
func StaticFS() (http.FileSystem, error) {
	frontendFS, err := fs.Sub(FS, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize static files: %w", err)
	}
	return http.FS(filteredFS{inner: frontendFS}), nil
}

// TemplateFS returns the filesystem for parsing HTML templates
func TemplateFS() fs.FS {
	return FS
}

// TemplateGlob returns all HTML template paths for parsing
func TemplateGlob() ([]string, error) {
	var templates []string
	err := fs.WalkDir(FS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".html" {
			templates = append(templates, path)
		}
		return nil
	})
	return templates, err
}
