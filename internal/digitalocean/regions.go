package digitalocean

// GetRegions returns all available DigitalOcean regions for Kubernetes clusters
func GetRegions() []string {
	// DigitalOcean regions that support Kubernetes
	return []string{
		"nyc1", "nyc3",
		"ams3",
		"fra1",
		"lon1",
		"sgp1",
		"tor1",
		"sfo2", "sfo3",
		"blr1",
		"syd1",
	}
}

// IsValidRegion checks if a region is valid for DigitalOcean Kubernetes
func IsValidRegion(region string) bool {
	regions := GetRegions()
	for _, r := range regions {
		if r == region {
			return true
		}
	}
	return false
}