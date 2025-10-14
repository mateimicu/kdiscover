package digitalocean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegions(t *testing.T) {
	regions := GetRegions()
	
	// Verify we have some regions
	assert.NotEmpty(t, regions)
	
	// Verify some expected regions are present
	expectedRegions := []string{"nyc1", "nyc3", "ams3", "fra1", "lon1", "sgp1", "sfo2", "sfo3", "tor1", "blr1", "syd1"}
	for _, expected := range expectedRegions {
		assert.Contains(t, regions, expected)
	}
}

func TestIsValidRegion(t *testing.T) {
	// Test valid regions
	validRegions := []string{"nyc1", "nyc3", "ams3", "fra1", "lon1", "sgp1", "sfo2", "sfo3", "tor1", "blr1", "syd1"}
	for _, region := range validRegions {
		assert.True(t, IsValidRegion(region), "Region %s should be valid", region)
	}
	
	// Test invalid regions
	invalidRegions := []string{"invalid", "us-east-1", "eu-central-1", ""}
	for _, region := range invalidRegions {
		assert.False(t, IsValidRegion(region), "Region %s should be invalid", region)
	}
}