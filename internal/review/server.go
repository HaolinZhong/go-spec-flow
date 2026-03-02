package review

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

// StartServer starts an HTTP server that serves the review HTML and accepts
// comment saves via POST /comments. It writes comments to review-comments.json
// in the specified directory. The server listens on the given port (0 for random).
func StartServer(tree *FlowTree, dir string, port int) (string, error) {
	mux := http.NewServeMux()

	// GET / — serve review HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := RenderHTML(tree, w, true); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// POST /comments — save comments to disk
	commentsPath := filepath.Join(dir, "review-comments.json")
	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var data json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Pretty-print the JSON
		formatted, err := json.MarshalIndent(json.RawMessage(data), "", "  ")
		if err != nil {
			formatted = data
		}

		if err := os.WriteFile(commentsPath, formatted, 0o644); err != nil {
			http.Error(w, "write failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})

	// Listen on the specified port
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("listen: %w", err)
	}

	// Extract port from the listener address
	tcpAddr := ln.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://localhost:%d", tcpAddr.Port)
	fmt.Fprintf(os.Stderr, "Review server running at %s\n", url)
	fmt.Fprintf(os.Stderr, "Comments will be saved to %s\n", commentsPath)
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")

	// Start serving (blocks until server stops)
	go http.Serve(ln, mux)

	return url, nil
}
