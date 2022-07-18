// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	cluster "github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update .golden files")

type simpleCluster struct {
	Endpoint                 string
	CertificateAuthorityData string
}

func TestLoadKubeconfig(t *testing.T) {
	cases := []struct {
		Clusters []*cluster.Cluster
	}{
		{Clusters: cluster.GetPredictableMockClusters(0)},
		{Clusters: cluster.GetPredictableMockClusters(1)},
		{Clusters: cluster.GetPredictableMockClusters(3)},
		{Clusters: cluster.GetPredictableMockClusters(10)},
	}

	for _, tc := range cases {
		tn := fmt.Sprintf("load_kubeconfig_%v", len(tc.Clusters))
		t.Run(tn, func(t *testing.T) {
			gp := filepath.Join("testdata", filepath.FromSlash(t.Name())+".golden")

			if *update {
				t.Log("update golden file")
				k := New()
				for _, c := range tc.Clusters {
					k.AddCluster(c, c.ID)
				}
				if err := k.Persist(gp); err != nil {
					t.Fatalf("failed to update golden file: %s", err)
				}
			}

			k, err := LoadKubeconfig(gp)
			if err != nil {
				t.Errorf("Failed to load kubeconfig %v", err.Error())
			}
			clusters, err := k.GetClusters()
			if err != nil {
				t.Errorf("Failed to load kubeconfig %v", err.Error())
			}

			out := make([]simpleCluster, 0)
			for _, c := range clusters {
				out = append(out, simpleCluster{
					Endpoint:                 c.Endpoint,
					CertificateAuthorityData: c.CertificateAuthorityData,
				})
			}

			expected := make([]simpleCluster, 0)
			for _, c := range tc.Clusters {
				expected = append(expected, simpleCluster{
					Endpoint:                 c.Endpoint,
					CertificateAuthorityData: c.CertificateAuthorityData,
				})
			}
			assert.ElementsMatch(t, out, expected)
		})
	}
}
