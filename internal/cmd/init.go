package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed all:embed_data
var embeddedFiles embed.FS

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install gsf skills and commands into the project",
	Long: `Installs gsf command files for Claude Code or Coco integration.
Detects whether the project uses .claude/ or .coco/ and installs accordingly.

Files are installed to <target>/commands/gsf/ so they become /gsf:<name> commands.
For example, gsf-review.md becomes /gsf:review.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := detectTargetDir()
		commandsDir := filepath.Join(targetDir, "commands", "gsf")
		fmt.Printf("Installing gsf commands to %s/\n", commandsDir)

		if err := os.MkdirAll(commandsDir, 0o755); err != nil {
			return fmt.Errorf("creating commands directory: %w", err)
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
			// Strip "gsf-" prefix: "gsf-review.md" → "review.md"
			destName := strings.TrimPrefix(entry.Name(), "gsf-")
			dest := filepath.Join(commandsDir, destName)
			if err := os.WriteFile(dest, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", dest, err)
			}
			cmdName := strings.TrimSuffix(destName, ".md")
			fmt.Printf("  /gsf:%s → %s\n", cmdName, dest)
		}

		fmt.Println("Done. Run /gsf:<command> to use.")
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
