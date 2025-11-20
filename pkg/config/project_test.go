package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProject(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test .project.toml file
	projectToml := filepath.Join(tmpDir, ".project.toml")
	content := `[project]
name = "Test Project"
id = "test-project"
status = "active"
type = "product"

[ownership]
primary = "test-owner"

[dates]
started = "2025-01-15"
`

	if err := os.WriteFile(projectToml, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load the project
	project, err := LoadProject(projectToml)
	if err != nil {
		t.Fatalf("LoadProject failed: %v", err)
	}

	// Verify project data
	if project.ProjectInfo.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", project.ProjectInfo.Name)
	}

	if project.ProjectInfo.ID != "test-project" {
		t.Errorf("Expected id 'test-project', got '%s'", project.ProjectInfo.ID)
	}

	if project.ProjectInfo.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", project.ProjectInfo.Status)
	}

	if project.Ownership.Primary != "test-owner" {
		t.Errorf("Expected owner 'test-owner', got '%s'", project.Ownership.Primary)
	}

	if project.Path != tmpDir {
		t.Errorf("Expected path '%s', got '%s'", tmpDir, project.Path)
	}
}

func TestLoadProjectMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	projectToml := filepath.Join(tmpDir, ".project.toml")

	// Create malformed TOML
	content := `[project
name = "Invalid`

	if err := os.WriteFile(projectToml, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should return error for malformed file
	_, err := LoadProject(projectToml)
	if err == nil {
		t.Error("Expected error for malformed TOML, got nil")
	}
}

func TestFindProjects(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create multiple projects
	projects := []struct {
		name string
		id   string
	}{
		{"Project One", "project-one"},
		{"Project Two", "project-two"},
	}

	for _, p := range projects {
		projectDir := filepath.Join(tmpDir, p.id)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}

		projectToml := filepath.Join(projectDir, ".project.toml")
		content := `[project]
name = "` + p.name + `"
id = "` + p.id + `"
status = "active"
type = "product"

[ownership]
primary = "test-owner"
`
		if err := os.WriteFile(projectToml, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create .project.toml: %v", err)
		}
	}

	// Find all projects
	foundProjects, err := FindProjects(tmpDir)
	if err != nil {
		t.Fatalf("FindProjects failed: %v", err)
	}

	// Verify we found both projects
	if len(foundProjects) != 2 {
		t.Errorf("Expected 2 projects, found %d", len(foundProjects))
	}

	// Verify project names
	foundIDs := make(map[string]bool)
	for _, p := range foundProjects {
		foundIDs[p.ProjectInfo.ID] = true
	}

	for _, p := range projects {
		if !foundIDs[p.id] {
			t.Errorf("Expected to find project '%s'", p.id)
		}
	}
}

func TestFindProjectsNonexistent(t *testing.T) {
	// Try to find projects in nonexistent directory
	projects, err := FindProjects("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("FindProjects should handle nonexistent dirs gracefully: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("Expected 0 projects from nonexistent dir, got %d", len(projects))
	}
}
