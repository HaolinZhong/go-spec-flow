package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/output"

	"github.com/spf13/cobra"
)

var (
	callersPkg  string
	callersFunc string
)

var callersCmd = &cobra.Command{
	Use:   "callers [dir]",
	Short: "Find direct callers of a function (one level)",
	Long: `Callers finds all functions that directly call the specified function.
Returns one level of callers only - the AI can decide whether to recurse.

Examples:
  gsf callers --pkg github.com/user/project/internal/review --func MapDiffToFunctions
  gsf callers --pkg sample-app/service --func CreateOrder --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if callersPkg == "" || callersFunc == "" {
			return fmt.Errorf("both --pkg and --func are required")
		}

		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		project, err := goast.LoadProject(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
			return err
		}

		result := goast.FindCallers(project, callersPkg, callersFunc)

		f := output.NewFormatter(format)
		return f.Print(result)
	},
}

func init() {
	callersCmd.Flags().StringVar(&callersPkg, "pkg", "", "package path of the target function")
	callersCmd.Flags().StringVar(&callersFunc, "func", "", "function name to find callers for")
	rootCmd.AddCommand(callersCmd)
}
