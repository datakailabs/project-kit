package session

import (
	"os"
	"testing"
)

func TestSanitizeSessionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"project-name", "project-name"},
		{"project.name", "project_name"},
		{"my.project.name", "my_project_name"},
		{"simple", "simple"},
		{"with.multiple.dots", "with_multiple_dots"},
	}

	for _, tt := range tests {
		result := SanitizeSessionName(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeSessionName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsInTmux(t *testing.T) {
	// Save original TMUX env var
	originalTmux := os.Getenv("TMUX")
	defer func() {
		if originalTmux != "" {
			os.Setenv("TMUX", originalTmux)
		} else {
			os.Unsetenv("TMUX")
		}
	}()

	// Test when TMUX is not set
	os.Unsetenv("TMUX")
	if IsInTmux() {
		t.Error("IsInTmux() should return false when TMUX is not set")
	}

	// Test when TMUX is set
	os.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	if !IsInTmux() {
		t.Error("IsInTmux() should return true when TMUX is set")
	}

	// Test when TMUX is empty string
	os.Setenv("TMUX", "")
	if IsInTmux() {
		t.Error("IsInTmux() should return false when TMUX is empty string")
	}
}
