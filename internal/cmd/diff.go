package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/output"
	"github.com/zhlie/go-spec-flow/internal/review"

	"github.com/spf13/cobra"
)

var (
	diffCommit          string
	diffBase            string
	diffIncludeUntracked bool
)

var diffCmd = &cobra.Command{
	Use:   "diff [dir]",
	Short: "Show function-level changes with complete code",
	Long: `Diff analyzes git changes and maps them to Go function/method declarations,
outputting each changed function with its full source code.

Without --commit or --base, shows staged changes (if any), otherwise unstaged changes.

Examples:
  gsf diff                                  # staged or unstaged changes
  gsf diff --commit HEAD                    # last commit
  gsf diff --base main                      # changes vs main branch
  gsf diff --include-untracked              # also include new untracked files
  gsf diff --format json                    # JSON output for AI consumption`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// Get diff with mode info
		diffOutput, err := review.GetDiffWithMode(dir, diffBase, diffCommit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
			return err
		}

		diffs := diffOutput.Diffs

		// Include untracked files if requested
		if diffIncludeUntracked {
			untracked, err := review.GetUntrackedGoFiles(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not get untracked files: %v\n", err)
			} else if len(untracked) > 0 {
				diffs = append(diffs, untracked...)
				fmt.Fprintf(os.Stderr, "Including %d untracked file(s)\n", len(untracked))
			}
		}

		goDiffs := review.FilterGoFiles(diffs)
		if len(goDiffs) == 0 {
			fmt.Println("No Go file changes found.")
			return nil
		}

		// Show diff mode hint for text output
		if format == "text" {
			switch diffOutput.Mode {
			case review.DiffModeStaged:
				fmt.Fprintf(os.Stderr, "Showing staged changes\n")
			case review.DiffModeUnstaged:
				fmt.Fprintf(os.Stderr, "Showing unstaged changes\n")
			}
		}

		// Load project for AST
		project, err := goast.LoadProject(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
			return err
		}

		// Map diff to functions
		changedFuncs := review.MapDiffToFunctions(goDiffs, project.RawPackages())
		if len(changedFuncs) == 0 {
			fmt.Println("No function-level changes detected.")
			return nil
		}

		// Extract full function code
		entries := review.ExtractDiffEntries(changedFuncs, project.RawPackages())
		result := &review.DiffResult{ChangedFunctions: entries}

		f := output.NewFormatter(format)
		return f.Print(result)
	},
}

func init() {
	diffCmd.Flags().StringVar(&diffCommit, "commit", "", "analyze a specific commit")
	diffCmd.Flags().StringVar(&diffBase, "base", "", "analyze changes against base branch")
	diffCmd.Flags().BoolVar(&diffIncludeUntracked, "include-untracked", false, "include untracked Go files as new")
	rootCmd.AddCommand(diffCmd)
}
