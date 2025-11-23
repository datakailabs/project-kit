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

var sessionsCmd = &cobra.Command{
	Use:   "sessions [name]",
	Short: "Switch between active tmux sessions (fast, Harpoon-style)",
	Long: `Switch between active tmux sessions quickly without filesystem scanning.

Unlike 'pk session' which shows ALL projects, 'pk sessions' only shows:
  - Currently running tmux sessions
  - No filesystem scanning (instant)
  - Perfect for quick switching between active work

If a project name is provided, switches directly to that session.
If no name is provided, shows an interactive fzf selector with active sessions only.

Bind this to Ctrl+b F (Shift+f) for fast access:
  bind-key F run-shell "tmux display-popup -E -w 90% -h 80% 'pk sessions'"

Examples:
  pk sessions           # Interactive picker (active sessions only)
  pk sessions pk        # Switch directly to 'pk' session`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return session.CheckTmux()
	},
	Run: runSessions,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
}

func runSessions(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")
	scriptoriumDir := filepath.Join(homeDir, "scriptorium")
	scratchDir := filepath.Join(homeDir, "scratch")

	// Get active tmux sessions
	activeSessions, err := session.ListSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list tmux sessions: %v\n", err)
		os.Exit(1)
	}

	if len(activeSessions) == 0 {
		fmt.Println("No active tmux sessions")
		fmt.Println("\nStart a session with:")
		fmt.Println("  pk session <project>")
		return
	}

	// Load all projects (from cache) to get metadata
	allProjects, err := cache.FindProjectsCached(projectsDir, archiveDir, scriptoriumDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load project metadata: %v\n", err)
		os.Exit(1)
	}

	// Also load scratch projects
	scratchProjects, _ := findScratchProjects(scratchDir)
	allProjects = append(allProjects, scratchProjects...)

	// Build map of active sessions to projects
	sessionProjects := make(map[string]*config.Project)
	for _, sessionName := range activeSessions {
		// Try to match session name to project
		for _, p := range allProjects {
			sanitizedID := session.SanitizeSessionName(p.ProjectInfo.ID)
			if sanitizedID == sessionName {
				sessionProjects[sessionName] = p
				break
			}
		}

		// If no project found, create a minimal one
		if _, found := sessionProjects[sessionName]; !found {
			sessionProjects[sessionName] = &config.Project{
				Path: "",
			}
			sessionProjects[sessionName].ProjectInfo.ID = sessionName
			sessionProjects[sessionName].ProjectInfo.Name = sessionName
			sessionProjects[sessionName].ProjectInfo.Status = "active"
		}
	}

	// If project name provided, switch directly
	if len(args) > 0 {
		targetName := strings.ToLower(args[0])
		targetSession := session.SanitizeSessionName(targetName)

		// Check if session exists
		found := false
		for sessionName := range sessionProjects {
			if sessionName == targetSession {
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Error: Session '%s' not found in active sessions\n", targetName)
			fmt.Fprintf(os.Stderr, "\nActive sessions:\n")
			for sessionName := range sessionProjects {
				fmt.Fprintf(os.Stderr, "  - %s\n", sessionName)
			}
			os.Exit(1)
		}

		// Switch to session
		if err := session.SwitchSession(targetSession); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to switch session: %v\n", err)
			os.Exit(1)
		}

		// Record access if we have project metadata
		if project, exists := sessionProjects[targetSession]; exists {
			cache.RecordAccess(project.ProjectInfo.ID, project.Path)
		}

		return
	}

	// Interactive selection with fzf
	selectedProject := selectActiveSessionWithFzf(sessionProjects)
	if selectedProject == nil {
		// User cancelled
		return
	}

	// Record access
	cache.RecordAccess(selectedProject.ProjectInfo.ID, selectedProject.Path)

	// Switch context if configured
	context.Switch(selectedProject)

	// Switch to session
	sessionName := session.SanitizeSessionName(selectedProject.ProjectInfo.ID)
	if err := session.SwitchSession(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to switch session: %v\n", err)
		os.Exit(1)
	}
}

func selectActiveSessionWithFzf(sessionProjects map[string]*config.Project) *config.Project {
	// Check if fzf is installed
	if _, err := exec.LookPath("fzf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: fzf is required for interactive selection\n")
		fmt.Fprintf(os.Stderr, "Install: brew install fzf (macOS) or apt install fzf (Linux)\n")
		fmt.Fprintf(os.Stderr, "\nAlternatively, specify a session: pk sessions <name>\n")
		os.Exit(1)
	}

	// Load pins to show which projects are pinned
	pins, _ := cache.ListPins()
	pinMap := make(map[string]int)
	for _, pin := range pins {
		pinMap[pin.ProjectID] = pin.Slot
	}

	// Build fzf input
	var builder strings.Builder
	projectMap := make(map[string]*config.Project)

	for _, p := range sessionProjects {
		owner := p.GetOwner()
		if owner == "" {
			owner = "none"
		}

		status := p.ProjectInfo.Status
		if status == "" {
			status = "active"
		}

		// Show pin number if pinned
		pinIndicator := ""
		if slot, isPinned := pinMap[p.ProjectInfo.ID]; isPinned {
			pinIndicator = fmt.Sprintf("[%d]", slot)
		}

		line := fmt.Sprintf("%s%s\t[%s]\t%s\t●\n",
			pinIndicator,
			p.ProjectInfo.ID,
			owner,
			status)
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
		"--prompt", "⚡ Active Session: ",
		"--preview", "echo 'Name: {1}\\nOwner: {2}\\nStatus: {3}\\nSession: {4}'",
		"--preview-window", "right:30%:wrap",
		"--header", "Active tmux sessions only | [N] = Pinned slot",
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

	// Get first field (project ID, potentially with [N] prefix)
	firstField := strings.Fields(selection)[0]

	// Remove pin indicator if present [1]pk -> pk
	projectID := firstField
	if strings.HasPrefix(projectID, "[") {
		// Extract ID after ]
		parts := strings.SplitN(projectID, "]", 2)
		if len(parts) == 2 {
			projectID = parts[1]
		}
	}

	return projectMap[projectID]
}
