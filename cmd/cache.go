package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/datakaicr/pk/pkg/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage project cache",
	Long: `Manage the project cache used by pk session for fast project discovery.

The cache is automatically maintained, but these commands allow manual control.

Subcommands:
  pk cache status    Show cache information
  pk cache refresh   Rebuild cache now
  pk cache clear     Remove cache file`,
}

var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache information",
	Run:   runCacheStatus,
}

var cacheRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Rebuild cache now",
	Run:   runCacheRefresh,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove cache file",
	Run:   runCacheClear,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
	cacheCmd.AddCommand(cacheRefreshCmd)
	cacheCmd.AddCommand(cacheClearCmd)
}

func runCacheStatus(cmd *cobra.Command, args []string) {
	status, err := cache.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(status)
}

func runCacheRefresh(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	projectsDir := filepath.Join(homeDir, "projects")
	archiveDir := filepath.Join(homeDir, "archive")
	scriptoriumDir := filepath.Join(homeDir, "scriptorium")

	fmt.Println("Refreshing cache...")

	// Clear old cache
	if err := cache.InvalidateCache(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not clear cache: %v\n", err)
	}

	// Rebuild cache
	cache.RebuildCacheAsync(projectsDir, archiveDir, scriptoriumDir)

	fmt.Println("\033[32m✓\033[0m Cache refresh triggered (rebuilding in background)")
	fmt.Println("\nRun 'pk cache status' to check progress")
}

func runCacheClear(cmd *cobra.Command, args []string) {
	if err := cache.InvalidateCache(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\033[32m✓\033[0m Cache cleared")
	fmt.Println("\nCache will be rebuilt on next 'pk session' or 'pk cache refresh'")
}
