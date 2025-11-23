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
	"github.com/spf13/cobra"
)

var (
	promoteMove   bool
	promoteNoGit  bool
	promoteOwner  string
	promoteType   string
)

var promoteCmd = &cobra.Command{
	Use:   "promote <path>",
	Short: "Convert directory into a project",
	Long: `Convert an existing directory into a proper project by:
  1. Moving it to ~/projects/<name> if --move is specified
  2. Creating .project.toml with metadata template
  3. Initializing git if not already a repository
  4. Auto-syncing shell aliases

Scratch projects in ~/scratch are automatically moved to ~/projects.

Example:
  pk promote api-test                            # Auto-detects scratch project
  pk promote /path/to/existing-work --move
  pk promote . --no-git                          # Promote current directory`,
	Args: cobra.ExactArgs(1),
	Run:  runPromote,
}

func init() {
	rootCmd.AddCommand(promoteCmd)
	promoteCmd.Flags().BoolVar(&promoteMove, "move", false,
		"Move to ~/projects/<name> (default: promote in place)")
	promoteCmd.Flags().BoolVar(&promoteNoGit, "no-git", false,
		"Skip git initialization if not already a repo")
	promoteCmd.Flags().StringVar(&promoteOwner, "owner", "datakai",
		"Project owner")
	promoteCmd.Flags().StringVar(&promoteType, "type", "product",
		"Project type")
}

func runPromote(cmd *cobra.Command, args []string) {
	// Get home directory first for scratch detection
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	// Resolve path - check if it's a scratch project name
	var dirPath string
	if args[0] == "." {
		dirPath, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not get current directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Check if it's a simple name (no path separators) - might be scratch project
		if !strings.Contains(args[0], string(filepath.Separator)) && !filepath.IsAbs(args[0]) {
			scratchPath := filepath.Join(homeDir, "scratch", args[0])
			if _, err := os.Stat(scratchPath); err == nil {
				dirPath = scratchPath
				promoteMove = true // Auto-enable move for scratch projects
				fmt.Printf("Detected scratch project: %s\n", scratchPath)
			} else {
				// Not in scratch, treat as relative path
				dirPath, err = filepath.Abs(args[0])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Invalid path: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			dirPath, err = filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid path: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Validate directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Directory not found: %s\n", dirPath)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Not a directory: %s\n", dirPath)
		os.Exit(1)
	}

	projectName := filepath.Base(dirPath)

	// Check if already a project
	tomlPath := filepath.Join(dirPath, ".project.toml")
	if _, err := os.Stat(tomlPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Already a project (found %s)\n", tomlPath)
		os.Exit(1)
	}

	// Move to ~/projects if --move
	if promoteMove {
		newPath := filepath.Join(homeDir, "projects", projectName)

		// Check if destination exists
		if _, err := os.Stat(newPath); err == nil {
			fmt.Fprintf(os.Stderr, "Error: Project already exists at %s\n", newPath)
			os.Exit(1)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create parent directory: %v\n", err)
			os.Exit(1)
		}

		// Move directory
		if err := os.Rename(dirPath, newPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to move directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Moved to: %s\n", newPath)
		dirPath = newPath
	}

	// Initialize git if needed
	gitDir := filepath.Join(dirPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		if !promoteNoGit {
			gitCmd := exec.Command("git", "init")
			gitCmd.Dir = dirPath
			if err := gitCmd.Run(); err != nil {
				fmt.Printf("Warning: git init failed: %v\n", err)
			} else {
				fmt.Println("Initialized git repository")
			}
		}
	} else {
		fmt.Println("Git repository already exists")
	}

	// Create .project.toml
	tomlPath = filepath.Join(dirPath, ".project.toml")
	if err := createPromoteProjectToml(tomlPath, projectName, dirPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create .project.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created metadata: %s\n", tomlPath)

	// Sync aliases
	fmt.Println("Syncing aliases...")
	runSync(cmd, []string{})

	fmt.Printf("\n\033[32m✓\033[0m Project '%s' promoted successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  %s      # Jump to project (after reloading shell)\n", projectName)
	fmt.Printf("  pk show %s\n", projectName)
}

func createPromoteProjectToml(path, name, projectPath string) error {
	// Create template project
	var project config.Project
	project.Path = projectPath
	project.ProjectInfo.Name = name
	project.ProjectInfo.ID = name
	project.ProjectInfo.Status = "active"
	project.ProjectInfo.Type = promoteType
	project.Consultant.Ownership = promoteOwner
	project.Consultant.MyRole = "owner"
	project.Tech.Stack = []string{}
	project.Tech.Domain = []string{}
	project.Dates.Started = time.Now().Format("2006-01-02")
	project.Dates.Completed = ""
	project.Links.Repository = ""
	project.Links.Documentation = ""
	project.Links.ConduitGraph = ""
	project.Notes.Description = ""

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
