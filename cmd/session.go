package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/datakaicr/pk/pkg/cache"
	"github.com/datakaicr/pk/pkg/config"
	"github.com/datakaicr/pk/pkg/context"
	"github.com/datakaicr/pk/pkg/session"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session [project]",
	Short: "Open project in tmux session (requires tmux)",
	Long: `Open a project in a tmux session with optional custom layouts.

If no project is specified, displays an interactive fzf selector.
If a project name is provided, opens that project directly.

Requires:
  - tmux: brew install tmux (macOS) or apt install tmux (Linux)
  - fzf: brew install fzf (macOS) or apt install fzf (Linux)

Custom layouts can be configured in .project.toml:

[tmux]
layout = "main-vertical"
windows = [
    {name = "editor", command = "nvim"},
    {name = "terminal"},
    {name = "server", command = "npm run dev"}
]

Example:
  pk session              # Interactive selector
  pk session dojo         # Open dojo project directly`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return session.CheckTmux()
	},
	Run:               runSession,
	ValidArgsFunction: validAllProjectNames,
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}

func runSession(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")
	scriptoriumDir := filepath.Join(homeDir, "scriptorium")
	scratchDir := filepath.Join(homeDir, "scratch")

	// Find all projects (uses cache if available)
	projects, err := cache.FindProjectsCached(projectsDir, archiveDir, scriptoriumDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find projects: %v\n", err)
		os.Exit(1)
	}

	// Also find scratch projects (no .project.toml required)
	scratchProjects, err := findScratchProjects(scratchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find scratch projects: %v\n", err)
		os.Exit(1)
	}

	// Combine projects and scratch
	allProjects := append(projects, scratchProjects...)

	var selectedProject *config.Project

	// If project name provided, find it directly
	if len(args) > 0 {
		projectName := strings.ToLower(args[0])
		for _, p := range allProjects {
			if strings.ToLower(p.ProjectInfo.ID) == projectName ||
				strings.ToLower(p.ProjectInfo.Name) == projectName {
				selectedProject = p
				break
			}
		}

		if selectedProject == nil {
			fmt.Fprintf(os.Stderr, "Error: Project '%s' not found\n", args[0])
			os.Exit(1)
		}
	} else {
		// Interactive selection with fzf
		selectedProject = selectProjectWithFzf(allProjects)
		if selectedProject == nil {
			// User cancelled
			return
		}
	}

	// Record project access
	cache.RecordAccess(selectedProject.ProjectInfo.ID, selectedProject.Path)

	// Switch context if configured
	context.Switch(selectedProject)

	// Create or switch to session
	if err := session.CreateSession(selectedProject); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create session: %v\n", err)
		os.Exit(1)
	}
}

// findScratchProjects finds directories in scratch (no .project.toml required)
func findScratchProjects(scratchDir string) ([]*config.Project, error) {
	var projects []*config.Project

	// Check if scratch directory exists
	if _, err := os.Stat(scratchDir); os.IsNotExist(err) {
		return projects, nil
	}

	// Read directories in scratch
	entries, err := os.ReadDir(scratchDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Create a pseudo-project for scratch directory
		scratchPath := filepath.Join(scratchDir, entry.Name())
		project := &config.Project{
			Path: scratchPath,
		}
		project.ProjectInfo.Name = entry.Name() + " (scratch)"
		project.ProjectInfo.ID = entry.Name()
		project.ProjectInfo.Status = "scratch"
		project.Consultant.Ownership = "scratch"

		projects = append(projects, project)
	}

	return projects, nil
}

func selectProjectWithFzf(projects []*config.Project) *config.Project {
	// Check if fzf is installed
	if _, err := exec.LookPath("fzf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: fzf is required for interactive selection\n")
		fmt.Fprintf(os.Stderr, "Install: brew install fzf (macOS) or apt install fzf (Linux)\n")
		fmt.Fprintf(os.Stderr, "\nAlternatively, specify a project: pk session <name>\n")
		os.Exit(1)
	}

	// Get list of existing sessions
	existingSessions, _ := session.ListSessions()
	sessionSet := make(map[string]bool)
	for _, s := range existingSessions {
		sessionSet[s] = true
	}

	// Build fzf input
	var builder strings.Builder
	projectMap := make(map[string]*config.Project)

	for _, p := range projects {
		// Format: "project-id    [owner]    status    [session-indicator]"
		owner := p.GetOwner()
		if owner == "" {
			owner = "none"
		}
		status := p.ProjectInfo.Status
		if status == "" {
			status = "unknown"
		}

		sessionName := session.SanitizeSessionName(p.ProjectInfo.ID)
		sessionIndicator := ""
		if sessionSet[sessionName] {
			sessionIndicator = "●" // Indicates active session
		}

		line := fmt.Sprintf("%s\t[%s]\t%s\t%s\n", p.ProjectInfo.ID, owner, status, sessionIndicator)
		builder.WriteString(line)
		projectMap[p.ProjectInfo.ID] = p
	}

	// Run fzf
	fzfCmd := exec.Command("fzf",
		"--height", "60%",
		"--reverse",
		"--border",
		"--ansi",
		"--tabstop=40",
		"--prompt", "⚡ Project: ",
		"--preview", "echo 'Name: {1}\\nOwner: {2}\\nStatus: {3}\\nSession: {4}'",
		"--preview-window", "right:30%:wrap",
		"--header", "● = Active Session",
	)

	fzfCmd.Stdin = strings.NewReader(builder.String())
	fzfCmd.Stderr = os.Stderr

	output, err := fzfCmd.Output()
	if err != nil {
		// User cancelled or error
		return nil
	}

	// Extract project ID from selection
	selection := strings.TrimSpace(string(output))
	if selection == "" {
		return nil
	}

	// Get first column (project ID)
	projectID := strings.Fields(selection)[0]
	return projectMap[projectID]
}
