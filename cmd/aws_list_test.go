// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"testing"

	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"

	"github.com/stretchr/testify/assert"
)

type mockExportable struct{}

func (mockExportable) IsExported(_ kubeconfig.Endpointer) bool {
	return false
}

type tableTestCase struct {
	Clusters []*cluster.Cluster
}

var (
	tableCases = []tableTestCase{
		{Clusters: cluster.GetMockClusters(0)},
		{Clusters: cluster.GetMockClusters(1)},
		{Clusters: cluster.GetMockClusters(3)},
	}
)

// Test if the number of Clusters is corectly diplayed
func Test_getTable(t *testing.T) {
	for _, tt := range tableCases {
		testname := fmt.Sprintf("Clusters %v", tt.Clusters)
		t.Run(testname, func(t *testing.T) {
			r := getTable(convertToInterfaces(tt.Clusters), mockExportable{}, "{{.Name}}-x")

			assert.Contains(t, r, fmt.Sprintf("%v", len(tt.Clusters)))

			for _, cls := range tt.Clusters {
				assert.Contains(t, r, fmt.Sprintf("%v-x", cls.GetName()))
			}
		})
	}
}

func Test_getTableBrokenTemplate(t *testing.T) {
	for _, tt := range tableCases {
		testname := fmt.Sprintf("Clusters %v", tt.Clusters)
		t.Run(testname, func(t *testing.T) {
			r := getTable(convertToInterfaces(tt.Clusters), mockExportable{}, "{{.NameX}}")

			assert.Contains(t, r, fmt.Sprintf("%v", len(tt.Clusters)))

			for _, cls := range tt.Clusters {
				assert.Contains(t, r, cls.GetName())
			}
		})
	}
}
