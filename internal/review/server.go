package review

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ServerConfig controls server behavior.
type ServerConfig struct {
	Port        int
	IdleTimeout time.Duration // 0 means no timeout
}

// StartServer starts an HTTP server that serves the review HTML and accepts
// comment saves via POST /comments. It writes comments to review-comments.json
// in the specified directory. Returns the URL and a channel that signals when
// the server should stop (via shutdown request or idle timeout).
func StartServer(tree *FlowTree, dir string, config ServerConfig) (string, <-chan struct{}, error) {
	mux := http.NewServeMux()
	done := make(chan struct{}, 1)

	// Idle timeout tracking
	var idleMu sync.Mutex
	lastActivity := time.Now()

	touch := func() {
		idleMu.Lock()
		lastActivity = time.Now()
		idleMu.Unlock()
	}

	// GET / — serve review HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		touch()
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
		touch()
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

	// POST /shutdown — graceful server shutdown
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
		select {
		case done <- struct{}{}:
		default:
		}
	})

	// Listen on the specified port
	addr := fmt.Sprintf(":%d", config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, fmt.Errorf("listen: %w", err)
	}

	// Extract port from the listener address
	tcpAddr := ln.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://localhost:%d", tcpAddr.Port)
	fmt.Fprintf(os.Stderr, "Review server running at %s\n", url)
	fmt.Fprintf(os.Stderr, "Comments will be saved to %s\n", commentsPath)
	if config.IdleTimeout > 0 {
		fmt.Fprintf(os.Stderr, "Auto-stop after %v of inactivity\n", config.IdleTimeout)
	}
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")

	// Start serving
	go http.Serve(ln, mux)

	// Idle timeout goroutine
	if config.IdleTimeout > 0 {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				idleMu.Lock()
				idle := time.Since(lastActivity)
				idleMu.Unlock()
				if idle >= config.IdleTimeout {
					fmt.Fprintf(os.Stderr, "\nServer idle for %v, stopping.\n", config.IdleTimeout)
					select {
					case done <- struct{}{}:
					default:
					}
					return
				}
			}
		}()
	}

	return url, done, nil
}
