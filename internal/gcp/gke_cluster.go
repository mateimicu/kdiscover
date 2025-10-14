// Package gcp provides function for working with GKE clusters
package gcp

import (
	"sync"

	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
)

type ClusterGetter interface {
	GetClusters(ch chan<- *cluster.Cluster)
}

func GetGKEClusters(projectsZones []ProjectZone) []*cluster.Cluster {
	clients := make([]ClusterGetter, 0, len(projectsZones))

	for _, pz := range projectsZones {
		log.WithFields(log.Fields{
			"project": pz.ProjectID,
			"zone":    pz.Zone,
		}).Info("Initialize client")
		gke, err := NewGKE(pz.ProjectID, pz.Zone)
		if err != nil {
			log.WithFields(log.Fields{
				"project": pz.ProjectID,
				"zone":    pz.Zone,
				"error":   err.Error(),
			}).Error("Failed to create GCP SDK session")
			continue
		}

		clients = append(clients, ClusterGetter(gke))
	}
	return getGKEClusters(clients)
}

// ProjectZone represents a GCP project and zone combination
type ProjectZone struct {
	ProjectID string
	Zone      string
}

// getGKEClusters will query the given project/zone combinations and return a list of
// clusters accessible. It will use the default credential chain for GCP
// in order to figure out the context for the API calls
func getGKEClusters(clients []ClusterGetter) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(clients))
	ch := make(chan *cluster.Cluster)

	var wg sync.WaitGroup
	wg.Add(len(clients))

	for _, c := range clients {
		regionCh := make(chan *cluster.Cluster)
		go c.GetClusters(regionCh)

		// fan-in from all the project/zones to one output channel
		go func(out chan<- *cluster.Cluster, wg *sync.WaitGroup) {
			for cls := range regionCh {
				out <- cls
			}
			wg.Done()
		}(ch, &wg)
	}

	// close the channel when all project/zones have finished the queries
	go func(wg *sync.WaitGroup, out chan<- *cluster.Cluster) {
		defer close(out)
		wg.Wait()
	}(&wg, ch)

	for cls := range ch {
		clusters = append(clusters, cls)
	}

	log.WithFields(log.Fields{
		"clusters-count": len(clusters),
	}).Info("Found clusters")

	return clusters
}