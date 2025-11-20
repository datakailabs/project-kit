package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/datakaicr/pk/pkg/config"
)

// AccessRecord tracks when a project was last accessed
type AccessRecord struct {
	ProjectID    string    `json:"project_id"`
	ProjectPath  string    `json:"project_path"`
	LastAccessed time.Time `json:"last_accessed"`
}

// GetAccessFile returns the path to the access tracking file
func GetAccessFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "access.json"), nil
}

// LoadAccessRecords reads the access tracking file
func LoadAccessRecords() (map[string]AccessRecord, error) {
	accessFile, err := GetAccessFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(accessFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]AccessRecord), nil
		}
		return nil, err
	}

	var records map[string]AccessRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// SaveAccessRecords writes the access tracking file
func SaveAccessRecords(records map[string]AccessRecord) error {
	accessFile, err := GetAccessFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(accessFile, data, 0644)
}

// RecordAccess marks a project as accessed now
func RecordAccess(projectID, projectPath string) error {
	records, err := LoadAccessRecords()
	if err != nil {
		return err
	}

	records[projectID] = AccessRecord{
		ProjectID:    projectID,
		ProjectPath:  projectPath,
		LastAccessed: time.Now(),
	}

	return SaveAccessRecords(records)
}

// GetRecentProjects returns projects sorted by access time (most recent first)
func GetRecentProjects(limit int) ([]*config.Project, error) {
	// Load access records
	records, err := LoadAccessRecords()
	if err != nil {
		return nil, err
	}

	// Load all projects
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projects, err := FindProjectsCached(
		filepath.Join(homeDir, "projects"),
		filepath.Join(homeDir, "scratch"),
	)
	if err != nil {
		return nil, err
	}

	// Sort projects by access time
	sort.Slice(projects, func(i, j int) bool {
		accessI, okI := records[projects[i].ProjectInfo.ID]
		accessJ, okJ := records[projects[j].ProjectInfo.ID]

		// Projects never accessed go to the end
		if !okI && !okJ {
			return false
		}
		if !okI {
			return false
		}
		if !okJ {
			return true
		}

		return accessI.LastAccessed.After(accessJ.LastAccessed)
	})

	// Apply limit
	if limit > 0 && limit < len(projects) {
		projects = projects[:limit]
	}

	return projects, nil
}
