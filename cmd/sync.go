package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/datakaicr/pk/pkg/config"
	"github.com/datakaicr/pk/pkg/shell"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync shell aliases for all projects",
	Long: `Generate shell aliases for all projects.

Detects your shell (zsh, bash, fish) and generates appropriate
alias files in the correct location.

For zsh: ~/.config/zsh/project-aliases.zsh
For bash: ~/.bash_aliases
For fish: ~/.config/fish/conf.d/project-aliases.fish

After running, reload your shell:
  source ~/.zshrc    # zsh
  source ~/.bashrc   # bash
  source ~/.config/fish/config.fish  # fish

Example:
  pk sync`,
	Run: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	// Detect shell
	currentShell := shell.Detect()
	fmt.Printf("Detected shell: \033[36m%s\033[0m\n", currentShell)

	// Find all projects
	homeDir, _ := os.UserHomeDir()
	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")

	fmt.Printf("Scanning projects...\n")
	projects, err := config.FindProjects(projectsDir, archiveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding projects: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	fmt.Printf("Found %d projects\n", len(projects))

	// Generate aliases
	fmt.Printf("Generating aliases...\n")
	if err := shell.GenerateAliases(currentShell, projects); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating aliases: %v\n", err)
		os.Exit(1)
	}

	aliasFile := shell.ConfigPath(currentShell)
	fmt.Printf("\n\033[32m✓\033[0m Aliases generated successfully!\n")
	fmt.Printf("  File: %s\n", aliasFile)
	fmt.Printf("\nReload your shell:\n")

	switch currentShell {
	case shell.Zsh:
		fmt.Printf("  source ~/.zshrc\n")
	case shell.Bash:
		fmt.Printf("  source ~/.bashrc\n")
	case shell.Fish:
		fmt.Printf("  source ~/.config/fish/config.fish\n")
	}

	fmt.Printf("\nThen test:\n")
	fmt.Printf("  dojo      # Jump to dojo project\n")
	fmt.Printf("  conduit   # Jump to conduit project\n")
}
