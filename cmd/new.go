package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/datakaicr/pk/pkg/config"
	"github.com/datakaicr/pk/pkg/hooks"
	"github.com/spf13/cobra"
)

var (
	newOwner string
	newType  string
	newNoGit bool
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new project",
	Long: `Create a new project with metadata template and optional git initialization.

This will:
  1. Create directory in ~/projects/<name>
  2. Initialize git repository (optional: --no-git)
  3. Create .project.toml with template metadata
  4. Auto-sync shell aliases

Example:
  pk new my-awesome-project
  pk new my-project --owner westmonroe --type client-project
  pk new prototype --no-git`,
	Args: cobra.ExactArgs(1),
	Run:  runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&newOwner, "owner", "datakai",
		"Project owner (datakai, westmonroe, etc.)")
	newCmd.Flags().StringVar(&newType, "type", "product",
		"Project type (product, client-project, internal)")
	newCmd.Flags().BoolVar(&newNoGit, "no-git", false,
		"Skip git initialization")
}

func runNew(cmd *cobra.Command, args []string) {
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

	projectPath := filepath.Join(homeDir, "projects", projectName)

	// Check if project already exists
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Project '%s' already exists at %s\n", projectName, projectPath)
		os.Exit(1)
	}

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create project directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created project directory: %s\n", projectPath)

	// Initialize git repository
	if !newNoGit {
		gitCmd := exec.Command("git", "init")
		gitCmd.Dir = projectPath
		if err := gitCmd.Run(); err != nil {
			fmt.Printf("Warning: git init failed: %v\n", err)
			fmt.Printf("Continuing without git...\n")
		} else {
			fmt.Println("Initialized git repository")
		}
	}

	// Create .project.toml
	tomlPath := filepath.Join(projectPath, ".project.toml")
	if err := createProjectToml(tomlPath, projectName, projectPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create .project.toml: %v\n", err)
		// Clean up
		os.RemoveAll(projectPath)
		os.Exit(1)
	}

	fmt.Printf("Created metadata: %s\n", tomlPath)

	// Sync aliases
	fmt.Println("Syncing aliases...")
	runSync(cmd, []string{})

	// Invalidate cache for pk session
	hooks.InvalidateCache()

	fmt.Printf("\n\033[32m✓\033[0m Project '%s' created successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd ~/projects/%s\n", projectName)
	fmt.Printf("  %s      # Jump to project (after reloading shell)\n", projectName)
}

func createProjectToml(path, name, projectPath string) error {
	// Create template project with NEW schema
	var project config.Project
	project.Path = projectPath

	// Core fields
	project.ProjectInfo.Name = name
	project.ProjectInfo.ID = name
	project.ProjectInfo.Status = "active"
	project.ProjectInfo.Type = newType
	project.Tech.Stack = []string{}
	project.Tech.Domain = []string{}
	project.Dates.Started = time.Now().Format("2006-01-02")
	project.Dates.Completed = ""
	project.Links.Repository = ""
	project.Links.Documentation = ""
	project.Notes.Description = ""

	// Consultant extension (only if owner is specified)
	if newOwner != "" {
		project.Consultant.Ownership = newOwner
		project.Consultant.MyRole = "owner"
	}

	// DataKai extension (only for DataKai projects)
	if newOwner == "datakai" {
		project.DataKai.Visibility = "private" // Default for new DataKai projects
	}

	// Write TOML file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	fmt.Fprintln(f, "# Project Metadata")
	fmt.Fprintln(f, "")

	encoder := toml.NewEncoder(f)
	return encoder.Encode(&project)
}
