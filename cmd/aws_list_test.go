// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mateimicu/kdiscover/internal/cluster"
)

type mockExportable struct{}

func (_ mockExportable) IsExported(cls *cluster.Cluster) bool {
	return false
}

// Test if the number of clusters is corectly diplayed
func Test_getTable(t *testing.T) {
	tts := []struct {
		clusters []cluster.Cluster
	}{
		{clusters: cluster.GetMockClusters(0)},
		{clusters: cluster.GetMockClusters(1)},
		{clusters: cluster.GetMockClusters(3)},
	}
	for _, tt := range tts {
		testname := fmt.Sprintf("Clusters %v", tt.clusters)
		t.Run(testname, func(t *testing.T) {
			r := getTable(tt.clusters, mockExportable{})
			if !strings.Contains(r, fmt.Sprintf("%v", len(tt.clusters))) {
				t.Errorf("Expected %v in output, but got %v", len(tt.clusters), r)
			}
		})
	}
}
