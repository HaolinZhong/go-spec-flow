package cmd

import (
	"fmt"
	"os"

	"github.com/zhlie/go-spec-flow/internal/output"
	"github.com/zhlie/go-spec-flow/internal/registry"

	"github.com/spf13/cobra"
)

var registryDir string

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage the cross-service RPC registry",
}

var registryUpdateCmd = &cobra.Command{
	Use:   "update <idl-dir>",
	Short: "Parse Thrift IDL files and update service registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idlDir := args[0]

		if err := os.MkdirAll(registryDir, 0o755); err != nil {
			return fmt.Errorf("creating registry directory: %w", err)
		}

		services, err := registry.GenerateFromIDL(idlDir, registryDir)
		if err != nil {
			return err
		}

		fmt.Printf("Updated registry with %d service(s):\n", len(services))
		for _, svc := range services {
			fmt.Printf("  %s (%d methods, from %s)\n", svc.Service, len(svc.Methods), svc.IDLPath)
		}
		return nil
	},
}

var registryShowCmd = &cobra.Command{
	Use:   "show <service-name>",
	Short: "Display service info (auto + context merged)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		info, err := registry.LoadServiceInfo(registryDir, serviceName)
		if err != nil {
			return fmt.Errorf("loading service info: %w", err)
		}

		ctx, _ := registry.LoadContext(registryDir, serviceName)
		merged := registry.MergeServiceInfo(info, ctx)

		f := output.NewFormatter(format)
		return f.Print(merged)
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered services",
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := registry.LoadIndex(registryDir)
		if err != nil {
			return fmt.Errorf("loading registry index: %w", err)
		}

		if len(idx.Services) == 0 {
			fmt.Println("No services registered. Run 'gsf registry update <idl-dir>' first.")
			return nil
		}

		f := output.NewFormatter(format)
		return f.Print(idx)
	},
}

func init() {
	registryCmd.PersistentFlags().StringVar(&registryDir, "registry-dir", "service-registry", "path to service registry directory")
	registryCmd.AddCommand(registryUpdateCmd)
	registryCmd.AddCommand(registryShowCmd)
	registryCmd.AddCommand(registryListCmd)
	rootCmd.AddCommand(registryCmd)
}
