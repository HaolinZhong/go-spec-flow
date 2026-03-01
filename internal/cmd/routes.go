package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/output"

	"github.com/spf13/cobra"
)

var routesCmd = &cobra.Command{
	Use:   "routes [dir]",
	Short: "Discover Hertz route registrations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		project, err := goast.LoadProject(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
			return err
		}

		rt := goast.DiscoverRoutes(project)
		if len(rt.Routes) == 0 {
			fmt.Println("No Hertz routes found.")
			return nil
		}

		f := output.NewFormatter(format)
		return f.Print(rt)
	},
}

func init() {
	rootCmd.AddCommand(routesCmd)
}
