package cmd

import (
	"fmt"
	"os"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/investigate"
	"github.com/zhlie/go-spec-flow/internal/output"
	"github.com/zhlie/go-spec-flow/internal/registry"

	"github.com/spf13/cobra"
)

// UnifiedContext combines all gsf analysis results into one document.
type UnifiedContext struct {
	ProjectStructure *goast.Project       `json:"project_structure" yaml:"project_structure"`
	Investigation    *investigate.Report  `json:"investigation" yaml:"investigation"`
	Registry         []*registry.MergedService `json:"service_registry,omitempty" yaml:"service_registry,omitempty"`
}

func (uc *UnifiedContext) String() string {
	s := ""
	if uc.ProjectStructure != nil {
		s += uc.ProjectStructure.String() + "\n"
	}
	if uc.Investigation != nil {
		s += uc.Investigation.String() + "\n"
	}
	if len(uc.Registry) > 0 {
		s += "Service Registry:\n"
		for _, svc := range uc.Registry {
			s += svc.String() + "\n"
		}
	}
	return s
}

var (
	ctxRoute     string
	ctxAllRoutes bool
	ctxRegistry  string
)

var contextCmd = &cobra.Command{
	Use:   "context [dir]",
	Short: "Generate unified AI-consumable context document",
	Long: `Combines project structure, investigation report, and service registry
into a single structured document for AI consumption.

Examples:
  gsf context --all-routes testdata/sample-app
  gsf context --route "POST /orders" testdata/sample-app`,
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

		uc := &UnifiedContext{
			ProjectStructure: project,
		}

		gen := investigate.NewGenerator(project, ctxRegistry)

		if ctxAllRoutes {
			report, err := gen.InvestigateAllRoutes()
			if err != nil {
				return err
			}
			uc.Investigation = report
		} else if ctxRoute != "" {
			report, err := gen.InvestigateRoutes([]string{ctxRoute})
			if err != nil {
				return err
			}
			uc.Investigation = report
		}

		// Load service registry if available
		if idx, err := registry.LoadIndex(ctxRegistry); err == nil {
			for _, svc := range idx.Services {
				info, err := registry.LoadServiceInfo(ctxRegistry, svc.Name)
				if err != nil {
					continue
				}
				ctx, _ := registry.LoadContext(ctxRegistry, svc.Name)
				uc.Registry = append(uc.Registry, registry.MergeServiceInfo(info, ctx))
			}
		}

		f := output.NewFormatter(format)
		return f.Print(uc)
	},
}

func init() {
	contextCmd.Flags().StringVar(&ctxRoute, "route", "", "route pattern to investigate")
	contextCmd.Flags().BoolVar(&ctxAllRoutes, "all-routes", false, "investigate all routes")
	contextCmd.Flags().StringVar(&ctxRegistry, "registry-dir", "service-registry", "service registry directory")
	rootCmd.AddCommand(contextCmd)
}
