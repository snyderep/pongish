package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type normalTemplateRenderer struct {
	templates map[string]*template.Template
}

func NewNormalTemplateRenderer(templateRoot string) *normalTemplateRenderer {
	templates := make(map[string]*template.Template)

	layouts, err := filepath.Glob(filepath.Join(templateRoot, "layouts", "*.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	includes, err := filepath.Glob(filepath.Join(templateRoot, "includes", "*.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, layout := range layouts {
		files := append(includes, layout)
		templates[filepath.Base(files[0])] = template.Must(template.ParseFiles(files...))
	}

	return &normalTemplateRenderer{templates: templates}
}

// renderTemplate is a wrapper around template.ExecuteTemplate.
func (r *normalTemplateRenderer) renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("The template '%s' does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}
