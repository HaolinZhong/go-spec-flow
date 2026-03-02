package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/review"

	"github.com/spf13/cobra"
)

var (
	reviewCommit   string
	reviewBase     string
	reviewCodebase bool
	reviewEntry    string
	reviewOutput   string
	reviewOpen     bool
	reviewJSON     bool
	reviewRender   string
	reviewDepth    int
	reviewServe    bool
)

var reviewCmd = &cobra.Command{
	Use:   "review [dir]",
	Short: "Generate an interactive HTML code review",
	Long: `Review generates a self-contained HTML file with an interactive flow tree
and code panel for reviewing code changes or exploring codebase architecture.

Diff mode (default): Shows git diff changes organized by file, with
function-level call chain context.

Codebase mode: Shows complete source code organized by call chains
from entry points.

JSON pipeline (for AI orchestration):
  gsf review --codebase --json > raw.json     # export structural data
  # AI edits raw.json: reorganizes flows, adds descriptions
  gsf review --render flow.json --open        # render AI-orchestrated HTML

Examples:
  gsf review --commit HEAD                          # last commit
  gsf review --commit HEAD~3..HEAD                  # last 3 commits
  gsf review --base main                            # vs main branch
  gsf review --codebase --entry "internal/ast"      # explore a package
  gsf review --codebase                             # all packages
  gsf review --commit HEAD --open                   # open in browser
  gsf review --commit HEAD --json                   # JSON to stdout
  gsf review --render flow.json --open              # render from JSON file`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// --render mode: read JSON and render HTML
		if reviewRender != "" {
			if reviewCommit != "" || reviewBase != "" || reviewCodebase {
				return fmt.Errorf("--render cannot be combined with --commit, --base, or --codebase")
			}
			return renderFromJSON(reviewRender)
		}

		// Build the flow tree
		var tree *review.FlowTree
		var err error

		if reviewCodebase {
			project, loadErr := goast.LoadProject(dir)
			if loadErr != nil {
				return fmt.Errorf("loading project: %w", loadErr)
			}
			tree, err = review.BuildCodebaseTree(project, reviewEntry, reviewDepth)
		} else {
			diffRange := buildDiffRange()
			if diffRange == "" {
				diffRange = "--staged"
			}
			tree, err = review.BuildDiffTree(dir, diffRange, reviewDepth)
		}

		if err != nil {
			return err
		}

		// --json mode: output JSON to stdout
		if reviewJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tree)
		}

		// --serve mode: start HTTP server
		if reviewServe {
			return serveReview(tree, dir)
		}

		// Default: render HTML
		return renderHTML(tree)
	},
}

func renderFromJSON(jsonFile string) error {
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("reading JSON file: %w", err)
	}

	var tree review.FlowTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	if reviewServe {
		return serveReview(&tree, ".")
	}
	return renderHTML(&tree)
}

func renderHTML(tree *review.FlowTree) error {
	output := reviewOutput
	if output == "" {
		output = "review.html"
	}

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := review.RenderHTML(tree, f, false); err != nil {
		return fmt.Errorf("rendering HTML: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Generated %s\n", output)

	if reviewOpen {
		openBrowser(output)
	}

	return nil
}

func serveReview(tree *review.FlowTree, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	config := review.ServerConfig{
		IdleTimeout: 30 * time.Minute,
	}
	url, done, err := review.StartServer(tree, absDir, config)
	if err != nil {
		return err
	}
	openBrowser(url)

	// Block until interrupted or server signals done (shutdown/idle)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
	case <-done:
	}
	fmt.Fprintln(os.Stderr, "\nServer stopped.")
	return nil
}

func buildDiffRange() string {
	if reviewCommit != "" {
		if contains(reviewCommit, "..") {
			return reviewCommit
		}
		return reviewCommit + "~1 " + reviewCommit
	}
	if reviewBase != "" {
		return reviewBase + "...HEAD"
	}
	return ""
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if isWSL() {
			cmd = exec.Command("cmd.exe", "/c", "start", path)
		} else {
			cmd = exec.Command("xdg-open", path)
		}
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		fmt.Fprintf(os.Stderr, "Open %s in your browser\n", path)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\nOpen %s manually\n", err, path)
	}
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return contains(string(data), "microsoft") || contains(string(data), "Microsoft")
}

func init() {
	reviewCmd.Flags().StringVar(&reviewCommit, "commit", "", "review a commit or range (e.g. HEAD, HEAD~3..HEAD)")
	reviewCmd.Flags().StringVar(&reviewBase, "base", "", "review changes against base branch")
	reviewCmd.Flags().BoolVar(&reviewCodebase, "codebase", false, "codebase exploration mode (instead of diff)")
	reviewCmd.Flags().StringVar(&reviewEntry, "entry", "", "entry package filter for codebase mode")
	reviewCmd.Flags().StringVar(&reviewOutput, "output", "", "output HTML file (default: review.html)")
	reviewCmd.Flags().BoolVar(&reviewOpen, "open", false, "open in browser after generating")
	reviewCmd.Flags().BoolVar(&reviewJSON, "json", false, "output FlowTree as JSON to stdout")
	reviewCmd.Flags().StringVar(&reviewRender, "render", "", "render HTML from a JSON file")
	reviewCmd.Flags().IntVar(&reviewDepth, "depth", 0, "max call chain trace depth (default: 2 for diff, 4 for codebase)")
	reviewCmd.Flags().BoolVar(&reviewServe, "serve", false, "start local server with comment support (auto-opens browser)")
	rootCmd.AddCommand(reviewCmd)
}
