package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/datakaicr/pk/pkg/config"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show detailed project information",
	Long: `Display detailed information about a specific project.

The project can be specified by its ID or name.

Example:
  pk show dojo
  pk show conduit
  pk show boardgamefinder`,
	Args:              cobra.ExactArgs(1),
	Run:               runShow,
	ValidArgsFunction: validProjectNames,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) {
	projectName := strings.ToLower(args[0])

	// Find projects
	homeDir, _ := os.UserHomeDir()
	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")

	projects, err := config.FindProjects(projectsDir, archiveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding projects: %v\n", err)
		os.Exit(1)
	}

	// Find matching project
	var found *config.Project
	for _, p := range projects {
		if strings.ToLower(p.ProjectInfo.ID) == projectName ||
			strings.ToLower(p.ProjectInfo.Name) == projectName {
			found = p
			break
		}
	}

	if found == nil {
		fmt.Fprintf(os.Stderr, "Project '%s' not found\n", projectName)
		os.Exit(1)
	}

	// Print detailed info
	printDetailedProject(found)
}

func printDetailedProject(p *config.Project) {
	// Header
	fmt.Printf("\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("  \033[1;34m%s\033[0m\n", p.ProjectInfo.Name)
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")

	// Project Info
	fmt.Printf("\033[1mProject Information\033[0m\n")
	fmt.Printf("  ID:          %s\n", p.ProjectInfo.ID)
	statusColor := getStatusColor(p.ProjectInfo.Status)
	fmt.Printf("  Status:      %s%s\033[0m\n", statusColor, p.ProjectInfo.Status)
	fmt.Printf("  Type:        %s\n", p.ProjectInfo.Type)
	fmt.Printf("  Path:        %s\n", p.Path)
	fmt.Printf("\n")

	// Ownership
	fmt.Printf("\033[1mOwnership\033[0m\n")
	fmt.Printf("  Owner:       %s\n", p.GetOwner())
	if partners := p.GetPartners(); len(partners) > 0 {
		fmt.Printf("  Partners:    %s\n", strings.Join(partners, ", "))
	}
	if p.GetLicenseModel() != "" {
		fmt.Printf("  License:     %s\n", p.GetLicenseModel())
	}
	fmt.Printf("\n")

	// Client Info (if applicable)
	if p.GetClientName() != "" || p.GetPartner() != "" {
		fmt.Printf("\033[1mClient Information\033[0m\n")
		if p.GetClientName() != "" {
			fmt.Printf("  End Client:  %s\n", p.GetClientName())
		}
		if p.GetPartner() != "" {
			fmt.Printf("  Via:         %s\n", p.GetPartner())
		}
		if p.GetMyRole() != "" {
			fmt.Printf("  Role:        %s\n", p.GetMyRole())
		}
		fmt.Printf("\n")
	}

	// Tech Stack
	if len(p.Tech.Stack) > 0 {
		fmt.Printf("\033[1mTechnology Stack\033[0m\n")
		fmt.Printf("  Stack:       %s\n", strings.Join(p.Tech.Stack, ", "))
		if len(p.Tech.Domain) > 0 {
			fmt.Printf("  Domain:      %s\n", strings.Join(p.Tech.Domain, ", "))
		}
		fmt.Printf("\n")
	}

	// Dates
	fmt.Printf("\033[1mTimeline\033[0m\n")
	fmt.Printf("  Started:     %s\n", p.Dates.Started)
	if p.Dates.Completed != "" {
		fmt.Printf("  Completed:   %s\n", p.Dates.Completed)
	} else {
		fmt.Printf("  Completed:   \033[32mOngoing\033[0m\n")
	}
	fmt.Printf("\n")

	// Links
	hasLinks := false
	linksStr := ""
	if p.Links.Repository != "" {
		linksStr += fmt.Sprintf("  Repository:  %s\n", p.Links.Repository)
		hasLinks = true
	}
	if p.Links.Documentation != "" {
		linksStr += fmt.Sprintf("  Docs:        %s\n", p.Links.Documentation)
		hasLinks = true
	}
	if p.Links.ScriptoriumProject != "" {
		linksStr += fmt.Sprintf("  Scriptorium: %s\n", p.Links.ScriptoriumProject)
		hasLinks = true
	}
	if hasLinks {
		fmt.Printf("\033[1mLinks\033[0m\n")
		fmt.Print(linksStr)
		fmt.Printf("\n")
	}

	// Description
	if p.Notes.Description != "" {
		fmt.Printf("\033[1mDescription\033[0m\n")
		fmt.Printf("  %s\n", p.Notes.Description)
		fmt.Printf("\n")
	}

	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")
}
