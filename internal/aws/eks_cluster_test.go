// Package aws provides function for working with EKS cluseters
package aws

import (
	"fmt"
	"sort"
	"testing"

	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/stretchr/testify/assert"
)

type fakeClusterGetter struct {
	Region   string
	Clusters []*cluster.Cluster
}

type sortedFakeClusterGetter []*cluster.Cluster

func (s sortedFakeClusterGetter) Len() int {
	return len(s)
}

func (s sortedFakeClusterGetter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedFakeClusterGetter) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func (c *fakeClusterGetter) GetClusters(ch chan<- *cluster.Cluster) {
	for _, cls := range c.Clusters {
		ch <- cls
	}
	close(ch)
}

func newFakeGetter(count int) *fakeClusterGetter {
	f := fakeClusterGetter{}
	f.Clusters = append(f.Clusters, cluster.GetMockClusters(count)...)
	return &f
}

func getAllClusters(clients []*fakeClusterGetter) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(clients))
	for _, c := range clients {
		clusters = append(clusters, c.Clusters...)
	}

	return clusters
}

func TestGetEKSClusters(t *testing.T) {
	t.Parallel()
	tts := []struct {
		Clients []*fakeClusterGetter
	}{
		{},
		{Clients: []*fakeClusterGetter{newFakeGetter(0)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(0)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(0), newFakeGetter(0)}},

		{Clients: []*fakeClusterGetter{newFakeGetter(1)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(1), newFakeGetter(1)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(1), newFakeGetter(1), newFakeGetter(1)}},

		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(1)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(1), newFakeGetter(1)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(9), newFakeGetter(1)}},

		{Clients: []*fakeClusterGetter{newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(10), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(10), newFakeGetter(10), newFakeGetter(10)}},

		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(10), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(0), newFakeGetter(10)}},

		{Clients: []*fakeClusterGetter{newFakeGetter(1), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(1), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(1), newFakeGetter(1), newFakeGetter(10)}},
		{Clients: []*fakeClusterGetter{newFakeGetter(0), newFakeGetter(0), newFakeGetter(10)}},

		{Clients: []*fakeClusterGetter{
			newFakeGetter(0),
			newFakeGetter(0),
			newFakeGetter(0),
			newFakeGetter(1),
			newFakeGetter(1),
			newFakeGetter(1),
			newFakeGetter(10),
			newFakeGetter(10),
			newFakeGetter(10),
			newFakeGetter(100),
			newFakeGetter(500),
			newFakeGetter(1000),
		}},
	}
	for _, tt := range tts {
		testname := fmt.Sprintf("Check all clusters are populated %v", len(tt.Clients))
		t.Run(testname, func(t *testing.T) {
			clients := make([]ClusterGetter, 0)
			for _, c := range tt.Clients {
				clients = append(clients, ClusterGetter(c))
			}
			allClusters := getAllClusters(tt.Clients)
			sort.Sort(sortedFakeClusterGetter(allClusters))

			r := getEKSClusters(clients)
			sort.Sort(sortedFakeClusterGetter(r))
			assert.Equal(t, r, allClusters)
		})
	}
}
