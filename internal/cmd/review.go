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
	reviewCommit string
	reviewBase   string
)

var reviewCmd = &cobra.Command{
	Use:   "review [dir]",
	Short: "Flow-based code review: see changes organized by request flow",
	Long: `Review analyzes git diff and overlays changes onto the call chain tree,
presenting code changes organized by request flow rather than by file.

Examples:
  gsf review testdata/sample-app                    # uncommitted changes
  gsf review --commit HEAD testdata/sample-app       # last commit
  gsf review --base main testdata/sample-app         # changes vs main branch`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// Get diff
		diffs, err := review.GetDiff(dir, reviewBase, reviewCommit)
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

		// Build flow review
		fr := review.BuildFlowReview(project, changedFuncs)

		f := output.NewFormatter(format)
		return f.Print(fr)
	},
}

func init() {
	reviewCmd.Flags().StringVar(&reviewCommit, "commit", "", "review a specific commit")
	reviewCmd.Flags().StringVar(&reviewBase, "base", "", "review changes against base branch")
	rootCmd.AddCommand(reviewCmd)
}
