package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/datakaicr/pk/pkg/cache"
	"github.com/spf13/cobra"
)

var cloneOpenSession bool

var cloneCmd = &cobra.Command{
	Use:   "clone <git-url> [name]",
	Short: "Clone a git repository and create .project.toml",
	Long: `Clone a git repository into ~/projects and automatically create a .project.toml file.

If the repository already contains a .project.toml, it will be preserved.
Otherwise, a basic configuration will be created.

The project name is extracted from the repository URL by default, but can
be overridden with the optional [name] argument.

Examples:
  pk clone https://github.com/user/repo
  pk clone git@github.com:user/repo.git
  pk clone https://github.com/user/repo my-project
  pk clone https://github.com/user/repo --session  # Open in tmux after cloning`,
	Args: cobra.MinimumNArgs(1),
	Run:  runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().BoolVarP(&cloneOpenSession, "session", "s", false, "Open in tmux session after cloning")
}

func runClone(cmd *cobra.Command, args []string) {
	gitURL := args[0]

	// Extract project name from URL
	projectName := extractProjectName(gitURL)

	// Override with provided name if given
	if len(args) > 1 {
		projectName = args[1]
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	projectsDir := filepath.Join(homeDir, "projects")
	targetPath := filepath.Join(projectsDir, projectName)

	// Check if project already exists
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Project '%s' already exists at %s\n", projectName, targetPath)
		os.Exit(1)
	}

	// Clone the repository
	fmt.Printf("Cloning %s into %s...\n", gitURL, targetPath)

	cloneCmd := exec.Command("git", "clone", gitURL, targetPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	if err := cloneCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to clone repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Repository cloned successfully")

	// Check if .project.toml already exists
	projectTomlPath := filepath.Join(targetPath, ".project.toml")
	if _, err := os.Stat(projectTomlPath); os.IsNotExist(err) {
		// Create a basic .project.toml
		if err := createBasicProjectToml(projectTomlPath, projectName, gitURL); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create .project.toml: %v\n", err)
		} else {
			fmt.Println("✓ Created .project.toml")
		}
	} else {
		fmt.Println("✓ Using existing .project.toml")
	}

	// Invalidate cache to pick up new project
	cache.InvalidateCache()

	fmt.Printf("\nProject '%s' ready at: %s\n", projectName, targetPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", targetPath)
	fmt.Printf("  pk edit %s          # Customize metadata\n", projectName)
	fmt.Printf("  pk session %s       # Open in tmux\n", projectName)

	// Open in session if requested
	if cloneOpenSession {
		fmt.Println("\nOpening in tmux session...")

		// Use the session command logic
		sessionArgs := []string{projectName}
		runSession(cmd, sessionArgs)
	}
}

// extractProjectName extracts the project name from a git URL
func extractProjectName(gitURL string) string {
	// Remove .git suffix if present
	url := strings.TrimSuffix(gitURL, ".git")

	// Handle different URL formats:
	// https://github.com/user/repo -> repo
	// git@github.com:user/repo -> repo
	// /path/to/repo -> repo

	// Split by / and get last part
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]

	// Handle SSH format (git@github.com:user/repo)
	if strings.Contains(name, ":") {
		parts := strings.Split(name, ":")
		name = parts[len(parts)-1]
	}

	// Sanitize name (lowercase, replace spaces/special chars)
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	return name
}

// createBasicProjectToml creates a minimal .project.toml file
func createBasicProjectToml(path, projectName, repoURL string) error {
	content := fmt.Sprintf(`# Project Metadata

[project]
name = "%s"
id = "%s"
status = "active"
type = "product"

[ownership]
primary = ""

[tech]
stack = []
domain = []

[dates]
started = "%s"
completed = ""

[links]
repository = "%s"
documentation = ""

[notes]
description = ""
`, projectName, projectName, getCurrentDate(), repoURL)

	return os.WriteFile(path, []byte(content), 0644)
}

// getCurrentDate returns the current date in YYYY-MM-DD format
func getCurrentDate() string {
	return time.Now().Format("2006-01-02")
}
