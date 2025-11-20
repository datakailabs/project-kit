package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/datakaicr/pk/pkg/session"
	"github.com/spf13/cobra"
)

var (
	scratchNoGit bool
)

var (
	scratchDeleteForce bool
)

var scratchCmd = &cobra.Command{
	Use:   "scratch",
	Short: "Manage scratch projects for experimentation",
	Long: `Manage temporary projects in ~/scratch for quick experiments.

Scratch projects are not tracked by pk (no .project.toml) and don't get
aliases. Use 'pk promote' to convert a scratch project into a real project.

Subcommands:
  pk scratch new <name>      Create a new scratch project
  pk scratch delete <name>   Delete a scratch project
  pk scratch list            List all scratch projects`,
}

var scratchNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new scratch project",
	Long: `Create a temporary project in ~/scratch for quick experiments.

This will:
  1. Create directory in ~/scratch/<name>
  2. Initialize git repository (optional: --no-git)
  3. Create basic README.md

Example:
  pk scratch new api-test              # Quick experiment
  pk scratch new prototype --no-git    # Without git

Then later:
  pk promote api-test`,
	Args: cobra.ExactArgs(1),
	Run:  runScratchNew,
}

var scratchDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a scratch project",
	Long: `Remove a scratch project from ~/scratch.

This will check for active tmux sessions and optionally kill them.

WARNING: This operation is permanent. Data will be deleted.

Example:
  pk scratch delete old-test
  pk scratch delete prototype --force  # Skip confirmation, auto-kill session`,
	Args:              cobra.ExactArgs(1),
	Run:               runScratchDelete,
	ValidArgsFunction: validScratchNames,
}

var scratchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scratch projects",
	Run:   runScratchList,
}

func init() {
	rootCmd.AddCommand(scratchCmd)
	scratchCmd.AddCommand(scratchNewCmd)
	scratchCmd.AddCommand(scratchDeleteCmd)
	scratchCmd.AddCommand(scratchListCmd)

	scratchNewCmd.Flags().BoolVar(&scratchNoGit, "no-git", false,
		"Skip git initialization")
	scratchDeleteCmd.Flags().BoolVar(&scratchDeleteForce, "force", false,
		"Skip confirmation prompt")
}

func runScratchNew(cmd *cobra.Command, args []string) {
	projectName := args[0]

	// Validate project name
	if strings.ContainsAny(projectName, "/\\:*?\"<>|") {
		fmt.Fprintf(os.Stderr, "Error: Invalid project name. Avoid special characters.\n")
		os.Exit(1)
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	scratchPath := filepath.Join(homeDir, "scratch", projectName)

	// Check if already exists
	if _, err := os.Stat(scratchPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Scratch project '%s' already exists at %s\n", projectName, scratchPath)
		os.Exit(1)
	}

	// Create directory
	if err := os.MkdirAll(scratchPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create scratch directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created scratch directory: %s\n", scratchPath)

	// Initialize git repository
	if !scratchNoGit {
		gitCmd := exec.Command("git", "init")
		gitCmd.Dir = scratchPath
		if err := gitCmd.Run(); err != nil {
			fmt.Printf("Warning: git init failed: %v\n", err)
		} else {
			fmt.Println("Initialized git repository")
		}
	}

	// Create basic README
	readmePath := filepath.Join(scratchPath, "README.md")
	readmeContent := fmt.Sprintf("# %s\n\nScratch project for experimentation.\n", projectName)
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		fmt.Printf("Warning: Failed to create README.md: %v\n", err)
	} else {
		fmt.Println("Created README.md")
	}

	fmt.Printf("\n\033[32m✓\033[0m Scratch project '%s' created!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd ~/scratch/%s\n", projectName)
	fmt.Printf("\nWhen ready to make it a real project:\n")
	fmt.Printf("  pk promote %s\n", projectName)
}

func runScratchDelete(cmd *cobra.Command, args []string) {
	projectName := args[0]

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	scratchPath := filepath.Join(homeDir, "scratch", projectName)

	// Check if exists
	if _, err := os.Stat(scratchPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Scratch project '%s' not found\n", projectName)
		os.Exit(1)
	}

	// Check for active tmux session
	sessionName := session.SanitizeSessionName(projectName)
	hasSession := session.SessionExists(sessionName)

	// Show confirmation prompt
	if !scratchDeleteForce {
		fmt.Printf("\033[33mWARNING: This will permanently delete the scratch project.\033[0m\n\n")
		fmt.Printf("Project:  %s\n", projectName)
		fmt.Printf("Location: %s\n", scratchPath)
		if hasSession {
			fmt.Printf("Tmux:     \033[33m● Active session found\033[0m\n")
		}
		fmt.Print("\nContinue? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" {
			fmt.Println("Cancelled")
			return
		}
	}

	// Kill tmux session if it exists
	if hasSession {
		if !scratchDeleteForce {
			fmt.Print("\nKill active tmux session? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" {
				if err := session.KillSession(sessionName); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to kill tmux session: %v\n", err)
				} else {
					fmt.Printf("\033[32m✓\033[0m Tmux session killed\n")
				}
			} else {
				fmt.Println("Tmux session will remain active")
			}
		} else {
			// Force flag: auto-kill session
			if err := session.KillSession(sessionName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to kill tmux session: %v\n", err)
			} else {
				fmt.Printf("\033[32m✓\033[0m Tmux session killed\n")
			}
		}
	}

	// Delete directory
	if err := os.RemoveAll(scratchPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to delete scratch project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\033[32m✓\033[0m Deleted: %s\n", scratchPath)
	fmt.Printf("\n\033[32m✓\033[0m Scratch project '%s' deleted successfully\n", projectName)
}

func runScratchList(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	scratchDir := filepath.Join(homeDir, "scratch")

	// Check if scratch directory exists
	if _, err := os.Stat(scratchDir); os.IsNotExist(err) {
		fmt.Println("No scratch directory found")
		return
	}

	// Read directories
	entries, err := os.ReadDir(scratchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read scratch directory: %v\n", err)
		os.Exit(1)
	}

	count := 0
	fmt.Println("=== Scratch Projects ===")
	fmt.Println()
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("\033[34m%s\033[0m\n", entry.Name())
			fmt.Printf("  Path: %s\n", filepath.Join(scratchDir, entry.Name()))
			fmt.Println()
			count++
		}
	}

	if count == 0 {
		fmt.Println("No scratch projects found")
	} else {
		fmt.Printf("Total: %d scratch projects\n", count)
	}
}
