package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordAccess(t *testing.T) {
	// Use a temporary cache directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	testHome := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	// Create cache directory
	cacheDir := filepath.Join(testHome, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Record an access
	projectID := "test-project"
	projectPath := "/path/to/project"

	err := RecordAccess(projectID, projectPath)
	if err != nil {
		t.Fatalf("RecordAccess failed: %v", err)
	}

	// Verify the record was saved
	records, err := LoadAccessRecords()
	if err != nil {
		t.Fatalf("LoadAccessRecords failed: %v", err)
	}

	record, exists := records[projectID]
	if !exists {
		t.Fatal("Access record not found")
	}

	if record.ProjectID != projectID {
		t.Errorf("Expected ProjectID '%s', got '%s'", projectID, record.ProjectID)
	}

	if record.ProjectPath != projectPath {
		t.Errorf("Expected ProjectPath '%s', got '%s'", projectPath, record.ProjectPath)
	}

	// Check that access time is recent (within last second)
	if time.Since(record.LastAccessed) > time.Second {
		t.Error("Access time is not recent")
	}
}

func TestLoadAccessRecordsEmpty(t *testing.T) {
	// Use a temporary cache directory with no access file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	testHome := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	// Create cache directory but no access.json file
	cacheDir := filepath.Join(testHome, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Should return empty map, not error
	records, err := LoadAccessRecords()
	if err != nil {
		t.Fatalf("LoadAccessRecords should handle missing file: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected empty records, got %d", len(records))
	}
}

func TestMultipleAccessRecords(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	testHome := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	cacheDir := filepath.Join(testHome, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Record multiple accesses
	projects := []struct {
		id   string
		path string
	}{
		{"project-1", "/path/to/project-1"},
		{"project-2", "/path/to/project-2"},
		{"project-3", "/path/to/project-3"},
	}

	for _, p := range projects {
		if err := RecordAccess(p.id, p.path); err != nil {
			t.Fatalf("RecordAccess failed for %s: %v", p.id, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Load and verify
	records, err := LoadAccessRecords()
	if err != nil {
		t.Fatalf("LoadAccessRecords failed: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	// Verify all projects are present
	for _, p := range projects {
		record, exists := records[p.id]
		if !exists {
			t.Errorf("Record for %s not found", p.id)
		}
		if record.ProjectPath != p.path {
			t.Errorf("Wrong path for %s: expected %s, got %s", p.id, p.path, record.ProjectPath)
		}
	}
}

func TestUpdateExistingAccess(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	testHome := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	cacheDir := filepath.Join(testHome, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	projectID := "test-project"
	projectPath := "/path/to/project"

	// First access
	if err := RecordAccess(projectID, projectPath); err != nil {
		t.Fatalf("First RecordAccess failed: %v", err)
	}

	records1, _ := LoadAccessRecords()
	firstAccess := records1[projectID].LastAccessed

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Second access
	if err := RecordAccess(projectID, projectPath); err != nil {
		t.Fatalf("Second RecordAccess failed: %v", err)
	}

	records2, _ := LoadAccessRecords()
	secondAccess := records2[projectID].LastAccessed

	// Second access should be later
	if !secondAccess.After(firstAccess) {
		t.Error("Second access time should be after first access time")
	}

	// Should still only have one record
	if len(records2) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records2))
	}
}
