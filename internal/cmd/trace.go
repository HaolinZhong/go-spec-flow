package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/output"

	"github.com/spf13/cobra"
)

var (
	traceDepth   int
	traceFunc    string
	tracePkg     string
	traceRoute   bool
)

var traceCmd = &cobra.Command{
	Use:   "trace [dir]",
	Short: "Trace call chains from an entry point",
	Long: `Trace builds a call chain tree from a given entry function.

Examples:
  # Trace from all discovered Hertz routes
  gsf trace --route testdata/sample-app

  # Trace from a specific function
  gsf trace --pkg sample-app/handler --func CreateOrder testdata/sample-app`,
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

		config := goast.TraceConfig{MaxDepth: traceDepth}
		tracer := goast.NewTracer(project, config)
		f := output.NewFormatter(format)

		if traceRoute {
			return traceFromRoutes(project, tracer, f)
		}

		if tracePkg != "" && traceFunc != "" {
			chain := tracer.Trace(tracePkg, traceFunc)
			return f.Print(chain)
		}

		return fmt.Errorf("specify --route to trace from routes, or --pkg and --func for a specific function")
	},
}

func traceFromRoutes(project *goast.Project, tracer *goast.Tracer, f *output.Formatter) error {
	rt := goast.DiscoverRoutes(project)
	if len(rt.Routes) == 0 {
		fmt.Println("No routes found.")
		return nil
	}

	for _, route := range rt.Routes {
		fmt.Printf("=== %s %s → %s ===\n", route.Method, route.Path, route.Handler)
		chain := tracer.TraceFromRoute(route)
		if err := f.Print(chain); err != nil {
			return err
		}
		fmt.Println()
	}
	return nil
}

func init() {
	traceCmd.Flags().IntVar(&traceDepth, "depth", 10, "maximum trace depth")
	traceCmd.Flags().StringVar(&traceFunc, "func", "", "function name to trace from")
	traceCmd.Flags().StringVar(&tracePkg, "pkg", "", "package path of the function")
	traceCmd.Flags().BoolVar(&traceRoute, "route", false, "trace from all discovered Hertz routes")
	rootCmd.AddCommand(traceCmd)
}
