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
	diffCommit string
	diffBase   string
)

var diffCmd = &cobra.Command{
	Use:   "diff [dir]",
	Short: "Show function-level changes with complete code",
	Long: `Diff analyzes git changes and maps them to Go function/method declarations,
outputting each changed function with its full source code.

Examples:
  gsf diff                                  # uncommitted changes
  gsf diff --commit HEAD                    # last commit
  gsf diff --base main                      # changes vs main branch
  gsf diff --format json                    # JSON output for AI consumption`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// Get diff
		diffs, err := review.GetDiff(dir, diffBase, diffCommit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
			return err
		}

		goDiffs := review.FilterGoFiles(diffs)
		if len(goDiffs) == 0 {
			fmt.Println("No Go file changes found.")
			return nil
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
	rootCmd.AddCommand(diffCmd)
}
