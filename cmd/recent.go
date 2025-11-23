package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/datakaicr/pk/pkg/cache"
	"github.com/spf13/cobra"
)

var recentLimit int

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "List recently accessed projects",
	Long: `List projects sorted by most recent access time.

Shows projects you've opened with 'pk session' recently. Projects never
accessed are not shown.

Examples:
  pk recent           # Show 10 most recent projects
  pk recent --limit 5 # Show 5 most recent projects`,
	Run: runRecent,
}

func init() {
	rootCmd.AddCommand(recentCmd)
	recentCmd.Flags().IntVarP(&recentLimit, "limit", "n", 10, "Number of projects to show")
}

func runRecent(cmd *cobra.Command, args []string) {
	projects, err := cache.GetRecentProjects(recentLimit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get recent projects: %v\n", err)
		os.Exit(1)
	}

	// Load access records to show times
	accessRecords, err := cache.LoadAccessRecords()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load access records: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No recently accessed projects")
		fmt.Println("\nTip: Projects are tracked when you open them with 'pk session'")
		return
	}

	fmt.Printf("Recently accessed projects (showing %d):\n\n", len(projects))

	for _, p := range projects {
		record, ok := accessRecords[p.ProjectInfo.ID]
		if !ok {
			continue
		}

		// Format access time
		accessTime := record.LastAccessed
		var timeStr string

		now := time.Now()
		diff := now.Sub(accessTime)

		if diff < time.Minute {
			timeStr = "just now"
		} else if diff < time.Hour {
			minutes := int(diff.Minutes())
			timeStr = fmt.Sprintf("%dm ago", minutes)
		} else if diff < 24*time.Hour {
			hours := int(diff.Hours())
			timeStr = fmt.Sprintf("%dh ago", hours)
		} else if diff < 7*24*time.Hour {
			days := int(diff.Hours() / 24)
			if days == 1 {
				timeStr = "1 day ago"
			} else {
				timeStr = fmt.Sprintf("%d days ago", days)
			}
		} else {
			timeStr = accessTime.Format("Jan 2, 2006")
		}

		owner := p.GetOwner()
		if owner == "" {
			owner = "none"
		}

		status := p.ProjectInfo.Status
		if status == "" {
			status = "unknown"
		}

		fmt.Printf("%-25s [%s] %-12s  %s\n",
			p.ProjectInfo.ID,
			owner,
			status,
			timeStr)
	}

	fmt.Printf("\nUse 'pk session <name>' to open a project\n")
}
