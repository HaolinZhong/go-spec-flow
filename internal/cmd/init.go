package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed all:embed_data
var embeddedFiles embed.FS

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install gsf skills and commands into the project",
	Long: `Installs gsf skill and command files for OpenSpec integration.
Detects whether the project uses .claude/ or .coco/ and installs accordingly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := detectTargetDir()
		fmt.Printf("Installing gsf skills to %s/\n", targetDir)

		// Install skills
		skillsDir := filepath.Join(targetDir, "skills")
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			return fmt.Errorf("creating skills directory: %w", err)
		}

		entries, err := fs.ReadDir(embeddedFiles, "embed_data/skills")
		if err != nil {
			return fmt.Errorf("reading embedded skills: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := embeddedFiles.ReadFile("embed_data/skills/" + entry.Name())
			if err != nil {
				return fmt.Errorf("reading %s: %w", entry.Name(), err)
			}
			dest := filepath.Join(skillsDir, entry.Name())
			if err := os.WriteFile(dest, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", dest, err)
			}
			fmt.Printf("  installed %s\n", dest)
		}

		fmt.Println("Done. gsf skills are ready for OpenSpec.")
		return nil
	},
}

func detectTargetDir() string {
	if _, err := os.Stat(".coco"); err == nil {
		return ".coco"
	}
	if _, err := os.Stat(".claude"); err == nil {
		return ".claude"
	}
	// Default to .claude
	return ".claude"
}

func init() {
	rootCmd.AddCommand(initCmd)
}
