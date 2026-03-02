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
// If serveMode is true, the HTML will auto-save comments to the server.
func RenderHTML(tree *FlowTree, w io.Writer, serveMode bool) error {
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
		"ServeMode":   serveMode,
	}

	return tmpl.Execute(w, data)
}
