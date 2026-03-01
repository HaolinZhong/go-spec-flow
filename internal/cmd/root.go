package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	format  string
)

var rootCmd = &cobra.Command{
	Use:   "gsf",
	Short: "Go Spec Flow - Static analysis engine for Go backend projects",
	Long: `gsf is a static analysis engine that provides structured code facts
(routes, call chains, RPC dependencies) for AI-driven development workflows.

It integrates with OpenSpec to enhance Go backend projects using Hertz and Kitex.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "text", "output format (text|json|yaml)")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gsf version %s\n", Version)
	},
}
