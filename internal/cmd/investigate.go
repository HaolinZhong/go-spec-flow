package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/investigate"
	"github.com/zhlie/go-spec-flow/internal/output"

	"github.com/spf13/cobra"
)

var (
	invRoute     string
	invPkg       string
	invFunc      string
	invAllRoutes bool
	invRegistry  string
)

var investigateCmd = &cobra.Command{
	Use:   "investigate [dir]",
	Short: "Generate investigation report for code context",
	Long: `Investigate analyzes code from specified entry points and generates
a structured report including call chains, modules, external dependencies,
and Service Registry cross-references.

Examples:
  gsf investigate --all-routes testdata/sample-app
  gsf investigate --route "POST /orders" testdata/sample-app
  gsf investigate --pkg sample-app/handler --func CreateOrder testdata/sample-app`,
	Args: cobra.MaximumNArgs(1),
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

		gen := investigate.NewGenerator(project, invRegistry)
		f := output.NewFormatter(format)

		if invAllRoutes {
			report, err := gen.InvestigateAllRoutes()
			if err != nil {
				return err
			}
			return f.Print(report)
		}

		if invRoute != "" {
			report, err := gen.InvestigateRoutes([]string{invRoute})
			if err != nil {
				return err
			}
			return f.Print(report)
		}

		if invPkg != "" && invFunc != "" {
			report, err := gen.InvestigateFunc(invPkg, invFunc)
			if err != nil {
				return err
			}
			return f.Print(report)
		}

		return fmt.Errorf("specify --all-routes, --route, or --pkg and --func")
	},
}

func init() {
	investigateCmd.Flags().StringVar(&invRoute, "route", "", "route pattern to investigate (e.g., 'POST /orders')")
	investigateCmd.Flags().StringVar(&invPkg, "pkg", "", "package path")
	investigateCmd.Flags().StringVar(&invFunc, "func", "", "function name")
	investigateCmd.Flags().BoolVar(&invAllRoutes, "all-routes", false, "investigate all discovered routes")
	investigateCmd.Flags().StringVar(&invRegistry, "registry-dir", "service-registry", "path to service registry directory")
	rootCmd.AddCommand(investigateCmd)
}
