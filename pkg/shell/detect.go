package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// Shell represents a detected shell type
type Shell string

const (
	Zsh  Shell = "zsh"
	Bash Shell = "bash"
	Fish Shell = "fish"
)

// Detect determines the current shell
func Detect() Shell {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		base := filepath.Base(shell)
		switch {
		case strings.Contains(base, "zsh"):
			return Zsh
		case strings.Contains(base, "bash"):
			return Bash
		case strings.Contains(base, "fish"):
			return Fish
		}
	}

	// Default to zsh (most common on macOS)
	return Zsh
}

// ConfigPath returns the path to the shell's alias config file
func ConfigPath(shell Shell) string {
	homeDir, _ := os.UserHomeDir()

	switch shell {
	case Zsh:
		// Use our custom project-aliases.zsh file
		return filepath.Join(homeDir, ".config", "zsh", "project-aliases.zsh")
	case Bash:
		return filepath.Join(homeDir, ".bash_aliases")
	case Fish:
		return filepath.Join(homeDir, ".config", "fish", "conf.d", "project-aliases.fish")
	default:
		return filepath.Join(homeDir, ".config", "zsh", "project-aliases.zsh")
	}
}

// String returns the shell name
func (s Shell) String() string {
	return string(s)
}
