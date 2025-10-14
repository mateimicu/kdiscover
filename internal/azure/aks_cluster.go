// Package azure provides function for working with AKS clusters
package azure

import (
	"sync"

	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
)

type ClusterGetter interface {
	GetClusters(ch chan<- *cluster.Cluster)
}

func GetAKSClusters(subscriptions []string) []*cluster.Cluster {
	clients := make([]ClusterGetter, 0, len(subscriptions))

	for _, subscription := range subscriptions {
		log.WithFields(log.Fields{
			"subscription": subscription,
		}).Info("Initialize AKS client")
		
		aks, err := NewAKS(subscription, "")
		if err != nil {
			log.WithFields(log.Fields{
				"subscription": subscription,
				"error":        err.Error(),
			}).Error("Failed to create Azure AKS client")
			continue
		}

		clients = append(clients, ClusterGetter(aks))
	}
	return getAKSClusters(clients)
}

// getAKSClusters will query the given clients and return a list of
// clusters accessible. It will use the default credential chain for Azure
// in order to figure out the context for the API calls
func getAKSClusters(clients []ClusterGetter) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(clients))
	ch := make(chan *cluster.Cluster)

	var wg sync.WaitGroup
	wg.Add(len(clients))

	for _, c := range clients {
		subscriptionCh := make(chan *cluster.Cluster)
		go c.GetClusters(subscriptionCh)

		// fan-in from all the subscriptions to one output channel
		go func(out chan<- *cluster.Cluster, wg *sync.WaitGroup) {
			for cls := range subscriptionCh {
				out <- cls
			}
			wg.Done()
		}(ch, &wg)
	}

	// close the channel when all subscriptions have finished the queries
	go func(wg *sync.WaitGroup, out chan<- *cluster.Cluster) {
		defer close(out)
		wg.Wait()
	}(&wg, ch)

	for cls := range ch {
		log.WithFields(log.Fields{
			"cluster": cls.Name,
			"region":  cls.Region,
		}).Debug("Found cluster")
		clusters = append(clusters, cls)
	}

	log.WithFields(log.Fields{
		"clusters": len(clusters),
	}).Info("Finished searching for AKS clusters")

	return clusters
}