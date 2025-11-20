package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/datakaicr/pk/pkg/shell"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install pk system-wide",
	Long: `Install pk binary, man page, and shell completions system-wide.

This command will:
  1. Create pk directories (~/projects, ~/scratch, ~/archive)
  2. Copy pk binary to /usr/local/bin/pk
  3. Install man page to system man directory
  4. Install shell completions for your shell
  5. Check for optional dependencies

Requires sudo permissions for binary and man page installation.

Core commands work without dependencies.
Optional features:
  - tmux session management: requires tmux and fzf
  - Context switching: requires cloud CLIs (aws, az, gcloud, etc.)

Example:
  pk install`,
	Run: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	fmt.Println("Installing PK (Project Kit)...")
	fmt.Println()

	// Get current binary location
	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine binary location: %v\n", err)
		os.Exit(1)
	}

	// Resolve symlinks
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not resolve binary path: %v\n", err)
		os.Exit(1)
	}

	projectRoot := filepath.Dir(filepath.Dir(binaryPath))
	manPagePath := filepath.Join(projectRoot, "docs", "pk.1")

	// Check if man page exists
	if _, err := os.Stat(manPagePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: Man page not found at %s\n", manPagePath)
		manPagePath = ""
	}

	// Get home directory for creating pk directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	// 1. Create pk directories
	fmt.Println("1. Creating pk directories...")
	projectsDir := filepath.Join(homeDir, "projects")
	scratchDir := filepath.Join(homeDir, "scratch")
	archiveDir := filepath.Join(homeDir, "archive")

	for _, dir := range []string{projectsDir, scratchDir, archiveDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to create %s: %v\n", dir, err)
		} else {
			if _, err := os.Stat(dir); err == nil {
				// Check if directory was just created or already existed
				fmt.Printf("   ✓ %s\n", dir)
			}
		}
	}
	fmt.Println()

	// 2. Install binary
	fmt.Println("2. Installing binary...")
	targetBinary := "/usr/local/bin/pk"

	cpCmd := exec.Command("sudo", "cp", binaryPath, targetBinary)
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	if err := cpCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to install binary: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   ✓ Binary installed to %s\n", targetBinary)
	fmt.Println()

	// 3. Install man page
	if manPagePath != "" {
		fmt.Println("3. Installing man page...")
		manDir := "/usr/local/share/man/man1"
		targetMan := filepath.Join(manDir, "pk.1")

		// Create man directory
		mkdirCmd := exec.Command("sudo", "mkdir", "-p", manDir)
		mkdirCmd.Run() // Ignore error if already exists

		// Copy man page
		cpManCmd := exec.Command("sudo", "cp", manPagePath, targetMan)
		cpManCmd.Stdout = os.Stdout
		cpManCmd.Stderr = os.Stderr
		if err := cpManCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to install man page: %v\n", err)
		} else {
			// Fix permissions
			chmodCmd := exec.Command("sudo", "chmod", "644", targetMan)
			chmodCmd.Run()

			fmt.Printf("   ✓ Man page installed to %s\n", targetMan)
		}
		fmt.Println()
	}

	// 4. Install shell completion
	fmt.Println("4. Installing shell completion...")
	installedCompletion := installCompletion()
	if installedCompletion {
		fmt.Println("   ✓ Shell completion installed")
	} else {
		fmt.Println("   ⚠ Shell completion not installed (unsupported shell)")
	}
	fmt.Println()

	// 5. Check optional dependencies
	fmt.Println("5. Checking optional dependencies...")
	checkDependency("tmux", "Required for 'pk session'")
	checkDependency("fzf", "Required for interactive 'pk session'")
	fmt.Println()

	// Success message
	fmt.Println("════════════════════════════════════════")
	fmt.Println("✓ Installation complete!")
	fmt.Println("════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Core commands available:")
	fmt.Println("  pk new/list/show/edit/delete/archive")
	fmt.Println("  pk scratch new/list/delete")
	fmt.Println("  pk sync")
	fmt.Println()
	if installedCompletion {
		fmt.Println("Shell completion installed. Reload your shell:")
		currentShell := shell.Detect()
		switch currentShell {
		case shell.Zsh:
			fmt.Println("  exec zsh")
		case shell.Bash:
			fmt.Println("  exec bash")
		case shell.Fish:
			fmt.Println("  exec fish")
		}
		fmt.Println()
	}
	fmt.Println("View documentation:")
	fmt.Println("  pk --help")
	if manPagePath != "" {
		fmt.Println("  man pk")
	}
}

func installCompletion() bool {
	currentShell := shell.Detect()

	switch currentShell {
	case shell.Zsh:
		return installZshCompletion()
	case shell.Bash:
		return installBashCompletion()
	case shell.Fish:
		return installFishCompletion()
	default:
		return false
	}
}

func installZshCompletion() bool {
	// Try Homebrew location first
	completionPath := ""
	if runtime.GOOS == "darwin" {
		brewPrefix, err := exec.Command("brew", "--prefix").Output()
		if err == nil {
			completionPath = filepath.Join(string(brewPrefix[:len(brewPrefix)-1]), "share", "zsh", "site-functions", "_pk")
		}
	}

	// Fallback to user directory
	if completionPath == "" {
		homeDir, _ := os.UserHomeDir()
		completionPath = filepath.Join(homeDir, ".zsh", "completions", "_pk")
		os.MkdirAll(filepath.Dir(completionPath), 0755)
	}

	// Generate completion
	compCmd := exec.Command("pk", "completion", "zsh")
	output, err := compCmd.Output()
	if err != nil {
		return false
	}

	// Write to file (use sudo if system directory)
	if filepath.HasPrefix(completionPath, "/usr") || filepath.HasPrefix(completionPath, "/opt") {
		// System directory - need sudo
		tmpFile := "/tmp/pk_completion.zsh"
		os.WriteFile(tmpFile, output, 0644)
		cpCmd := exec.Command("sudo", "cp", tmpFile, completionPath)
		return cpCmd.Run() == nil
	} else {
		// User directory
		return os.WriteFile(completionPath, output, 0644) == nil
	}
}

func installBashCompletion() bool {
	homeDir, _ := os.UserHomeDir()
	completionDir := filepath.Join(homeDir, ".bash_completion.d")
	os.MkdirAll(completionDir, 0755)

	completionPath := filepath.Join(completionDir, "pk")

	compCmd := exec.Command("pk", "completion", "bash")
	output, err := compCmd.Output()
	if err != nil {
		return false
	}

	return os.WriteFile(completionPath, output, 0644) == nil
}

func installFishCompletion() bool {
	homeDir, _ := os.UserHomeDir()
	completionDir := filepath.Join(homeDir, ".config", "fish", "completions")
	os.MkdirAll(completionDir, 0755)

	completionPath := filepath.Join(completionDir, "pk.fish")

	compCmd := exec.Command("pk", "completion", "fish")
	output, err := compCmd.Output()
	if err != nil {
		return false
	}

	return os.WriteFile(completionPath, output, 0644) == nil
}

func checkDependency(name, description string) {
	if _, err := exec.LookPath(name); err == nil {
		fmt.Printf("   ✓ %s installed\n", name)
	} else {
		fmt.Printf("   ⚠ %s not found - %s\n", name, description)
	}
}
