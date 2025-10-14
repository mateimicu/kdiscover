// Package azure provides function for working with AKS clusters
package azure

import (
	"fmt"
	"testing"

	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/stretchr/testify/assert"
)

type fakeAKSClusterGetter struct {
	SubscriptionID string
	Clusters       []*cluster.Cluster
}

func (c *fakeAKSClusterGetter) GetClusters(ch chan<- *cluster.Cluster) {
	for _, cls := range c.Clusters {
		ch <- cls
	}
	close(ch)
}

func newFakeAKSGetter(count int) *fakeAKSClusterGetter {
	f := fakeAKSClusterGetter{}
	f.Clusters = append(f.Clusters, cluster.GetMockClusters(count)...)
	// Set provider to Azure for all mock clusters
	for _, cls := range f.Clusters {
		cls.Provider = cluster.Azure
	}
	return &f
}

func getAllAKSClusters(clients []*fakeAKSClusterGetter) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(clients))
	for _, c := range clients {
		clusters = append(clusters, c.Clusters...)
	}

	return clusters
}

func TestGetAKSClusters(t *testing.T) {
	t.Parallel()
	tts := []struct {
		Clients []*fakeAKSClusterGetter
	}{
		{Clients: []*fakeAKSClusterGetter{}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(0)}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(1)}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(10)}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(0), newFakeAKSGetter(1), newFakeAKSGetter(10)}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(1), newFakeAKSGetter(1), newFakeAKSGetter(10)}},
		{Clients: []*fakeAKSClusterGetter{newFakeAKSGetter(0), newFakeAKSGetter(0), newFakeAKSGetter(10)}},

		{Clients: []*fakeAKSClusterGetter{
			newFakeAKSGetter(0),
			newFakeAKSGetter(0),
			newFakeAKSGetter(0),
			newFakeAKSGetter(1),
			newFakeAKSGetter(1),
			newFakeAKSGetter(1),
			newFakeAKSGetter(10),
			newFakeAKSGetter(10),
			newFakeAKSGetter(10),
			newFakeAKSGetter(100),
			newFakeAKSGetter(500),
			newFakeAKSGetter(1000),
		}},
	}
	for _, tt := range tts {
		testname := fmt.Sprintf("Check all clusters are populated %v", len(tt.Clients))
		t.Run(testname, func(t *testing.T) {
			clients := make([]ClusterGetter, 0)
			for _, c := range tt.Clients {
				clients = append(clients, ClusterGetter(c))
			}
			allClusters := getAllAKSClusters(tt.Clients)

			r := getAKSClusters(clients)
			assert.ElementsMatch(t, r, allClusters)
		})
	}
}