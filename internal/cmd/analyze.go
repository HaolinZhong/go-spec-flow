package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/output"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [dir]",
	Short: "Analyze project structure (packages, structs, interfaces, functions)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		project, err := goast.LoadProject(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		f := output.NewFormatter(format)
		return f.Print(project)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
