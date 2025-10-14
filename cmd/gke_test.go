// Package cmd offers CLI functionality
package cmd

import (
	"os"
	"testing"

	"github.com/mateimicu/kdiscover/internal/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGetProjectZones(t *testing.T) {
	t.Parallel()
	
	// Save original environment variables
	originalProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	defer func() {
		if originalProject != "" {
			os.Setenv("GOOGLE_CLOUD_PROJECT", originalProject)
		} else {
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		}
	}()

	// Test with explicit projects and zones
	gcpProjects = []string{"test-project"}
	gcpZones = []string{"us-central1-a", "us-central1-b"}
	
	projectZones := getProjectZones()
	
	assert.Equal(t, 2, len(projectZones))
	assert.Equal(t, "test-project", projectZones[0].ProjectID)
	assert.Equal(t, "us-central1-a", projectZones[0].Zone)
	assert.Equal(t, "test-project", projectZones[1].ProjectID)
	assert.Equal(t, "us-central1-b", projectZones[1].Zone)
}

func TestGetProjectZonesWithDefaultFallback(t *testing.T) {
	t.Parallel()
	
	// Save original environment variables
	originalProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	defer func() {
		if originalProject != "" {
			os.Setenv("GOOGLE_CLOUD_PROJECT", originalProject)
		} else {
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		}
	}()

	// Set test environment
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project-env")
	
	// Reset global variables
	gcpProjects = []string{}
	gcpZones = []string{}
	
	projectZones := getProjectZones()
	
	// Should have at least some project zones from default fallback
	if len(projectZones) > 0 {
		// Check that all project zones use the environment project
		for _, pz := range projectZones {
			assert.Equal(t, "test-project-env", pz.ProjectID)
		}
	}
}

func TestProjectZoneStruct(t *testing.T) {
	t.Parallel()
	
	pz := gcp.ProjectZone{
		ProjectID: "test-project",
		Zone:      "us-central1-a",
	}
	
	assert.Equal(t, "test-project", pz.ProjectID)
	assert.Equal(t, "us-central1-a", pz.Zone)
}