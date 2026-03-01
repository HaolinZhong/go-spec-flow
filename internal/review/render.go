package review

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
)

//go:embed templates/review.html
var templateFS embed.FS

// RenderHTML renders a FlowTree as a self-contained HTML file.
func RenderHTML(tree *FlowTree, w io.Writer) error {
	tmplContent, err := templateFS.ReadFile("templates/review.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("review").Parse(string(tmplContent))
	if err != nil {
		return err
	}

	treeJSON, err := json.Marshal(tree)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":       tree.Title,
		"Mode":        tree.Mode,
		"Description": tree.Description,
		"TreeJSON":    template.JS(treeJSON),
	}

	return tmpl.Execute(w, data)
}
