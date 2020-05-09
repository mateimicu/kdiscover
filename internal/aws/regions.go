package aws

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

func contains(key string, list []string) bool {
	for _, val := range list {
		if key == val {
			return true
		}
	}
	return false
}

// GetRegions will return all the available regions in the given
// aws partitions
func GetRegions(awsPartitions []string) []string {
	var regions []string
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		if !contains(p.ID(), awsPartitions) {
			continue
		}

		for id := range p.Regions() {
			regions = append(regions, id)
		}
	}

	return regions
}

// AllowedParitions returns all the allowed AWS partitions
func AllowedParitions() []string {
	var allowedPartitions []string
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		allowedPartitions = append(allowedPartitions, p.ID())
	}

	return allowedPartitions
}
