package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/datakaicr/pk/pkg/config"
)

const (
	// Cache configuration
	CacheMaxAge = 5 * time.Minute // Refresh every 5 minutes (shorter than bash!)
)

// GetCacheFile returns the path to the cache file
func GetCacheFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "pk")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "projects.json"), nil
}

// IsCacheValid checks if cache exists and is recent
func IsCacheValid() bool {
	cacheFile, err := GetCacheFile()
	if err != nil {
		return false
	}

	info, err := os.Stat(cacheFile)
	if err != nil {
		return false
	}

	age := time.Since(info.ModTime())
	return age < CacheMaxAge
}

// LoadFromCache reads projects from cache
func LoadFromCache() ([]*config.Project, error) {
	cacheFile, err := GetCacheFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var projects []*config.Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// SaveToCache writes projects to cache
func SaveToCache(projects []*config.Project) error {
	cacheFile, err := GetCacheFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// FindProjectsCached returns projects from cache if valid, otherwise scans and caches
func FindProjectsCached(rootDirs ...string) ([]*config.Project, error) {
	// Try cache first
	if IsCacheValid() {
		projects, err := LoadFromCache()
		if err == nil {
			return projects, nil
		}
		// Cache read failed, fall through to scan
	}

	// Scan filesystem
	projects, err := config.FindProjects(rootDirs...)
	if err != nil {
		return nil, err
	}

	// Update cache in background (non-blocking)
	go func() {
		SaveToCache(projects)
	}()

	return projects, nil
}

// InvalidateCache removes the cache file
func InvalidateCache() error {
	cacheFile, err := GetCacheFile()
	if err != nil {
		return err
	}

	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// RebuildCacheAsync triggers a cache rebuild in the background
func RebuildCacheAsync(rootDirs ...string) {
	go func() {
		InvalidateCache()
		projects, err := config.FindProjects(rootDirs...)
		if err == nil {
			SaveToCache(projects)
		}
	}()
}

// Status returns cache information
func Status() (string, error) {
	cacheFile, err := GetCacheFile()
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "Cache: not built", nil
		}
		return "", err
	}

	age := time.Since(info.ModTime())
	valid := age < CacheMaxAge

	status := fmt.Sprintf("Cache: %s\n", cacheFile)
	status += fmt.Sprintf("Age: %s\n", age.Round(time.Second))
	status += fmt.Sprintf("Valid: %v\n", valid)
	status += fmt.Sprintf("Size: %d bytes\n", info.Size())

	return status, nil
}
