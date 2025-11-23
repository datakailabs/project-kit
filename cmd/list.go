package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/datakaicr/pk/pkg/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [filter]",
	Short: "List all projects",
	Long: `List all projects with optional filter.

Filters:
  active      - Only active projects
  archived    - Only archived projects
  datakai     - DataKai projects
  westmonroe  - West Monroe projects
  product     - Product projects
  client      - Client projects

Examples:
  pk list              # All projects
  pk list active       # Active projects only
  pk list datakai      # DataKai projects only`,
	Run:               runList,
	ValidArgsFunction: validListFilters,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	// Get filter (if provided)
	filter := ""
	if len(args) > 0 {
		filter = strings.ToLower(args[0])
	}

	// Find projects in standard locations
	homeDir, _ := os.UserHomeDir()
	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")

	projects, err := config.FindProjects(projectsDir, archiveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding projects: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	// Apply filter
	filtered := filterProjects(projects, filter)

	// Print header
	fmt.Printf("\n=== Projects (%s) ===\n\n", getFilterLabel(filter))

	// Print each project
	for _, p := range filtered {
		printProject(p)
	}

	fmt.Printf("\nTotal: %d projects\n", len(filtered))
}

func filterProjects(projects []*config.Project, filter string) []*config.Project {
	if filter == "" {
		return projects
	}

	var filtered []*config.Project
	for _, p := range projects {
		switch filter {
		case "active":
			if p.ProjectInfo.Status == "active" {
				filtered = append(filtered, p)
			}
		case "archived":
			if p.ProjectInfo.Status == "archived" {
				filtered = append(filtered, p)
			}
		case "datakai":
			if p.GetOwner() == "datakai" {
				filtered = append(filtered, p)
			}
		case "westmonroe":
			if p.GetOwner() == "westmonroe" {
				filtered = append(filtered, p)
			}
		case "product":
			if p.ProjectInfo.Type == "product" {
				filtered = append(filtered, p)
			}
		case "client", "client-project":
			if p.ProjectInfo.Type == "client-project" {
				filtered = append(filtered, p)
			}
		}
	}
	return filtered
}

func getFilterLabel(filter string) string {
	if filter == "" {
		return "all"
	}
	return filter
}

func printProject(p *config.Project) {
	// Project name and ID
	fmt.Printf("\033[34m%s\033[0m\n", p.ProjectInfo.ID)
	fmt.Printf("  Name: %s\n", p.ProjectInfo.Name)

	// Status (with color)
	statusColor := getStatusColor(p.ProjectInfo.Status)
	fmt.Printf("  Status: %s%s\033[0m | Type: %s | Owner: %s\n",
		statusColor,
		p.ProjectInfo.Status,
		p.ProjectInfo.Type,
		p.GetOwner())

	// Path
	fmt.Printf("  Path: %s\n", p.Path)

	fmt.Println()
}

func getStatusColor(status string) string {
	switch status {
	case "active":
		return "\033[32m" // Green
	case "archived":
		return "\033[33m" // Yellow
	case "paused":
		return "\033[36m" // Cyan
	default:
		return ""
	}
}
