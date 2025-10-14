// Package digitalocean provides function for working with DOKS clusters
package digitalocean

import (
	"sync"

	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientAPIVersion = "client.authentication.k8s.io/v1beta1"
)

func getConfigAuthInfo(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
	authInfo := clientcmdapi.NewAuthInfo()
	
	// DigitalOcean uses doctl for authentication
	authInfo.Exec = &clientcmdapi.ExecConfig{
		Command: "doctl",
		Args: []string{
			"kubernetes",
			"cluster",
			"kubeconfig",
			"exec-credential",
			"--cluster-id", cls.ID,
		},
		APIVersion: clientAPIVersion,
	}
	return authInfo
}

type ClusterGetter interface {
	GetClusters(ch chan<- *cluster.Cluster)
}

// GetDOKSClusters retrieves all DOKS clusters from specified regions
func GetDOKSClusters(regions []string) []*cluster.Cluster {
	clients := make([]ClusterGetter, 0, len(regions))

	for _, region := range regions {
		log.WithFields(log.Fields{
			"region": region,
		}).Info("Initialize DOKS client")
		doks, err := NewDOKS(region)
		if err != nil {
			log.WithFields(log.Fields{
				"region": region,
				"error":  err.Error(),
			}).Error("Failed to create DigitalOcean client")
			continue
		}

		clients = append(clients, ClusterGetter(doks))
	}
	return getDOKSClusters(clients)
}

// getDOKSClusters will query the given clients and return a list of
// clusters accessible. It will use the DIGITALOCEAN_TOKEN environment variable
// for authentication
func getDOKSClusters(clients []ClusterGetter) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(clients))
	ch := make(chan *cluster.Cluster)

	var wg sync.WaitGroup
	wg.Add(len(clients))

	for _, c := range clients {
		regionCh := make(chan *cluster.Cluster)
		go c.GetClusters(regionCh)

		// fan-in from all the regions to one output channel
		go func(out chan<- *cluster.Cluster, wg *sync.WaitGroup) {
			for cls := range regionCh {
				out <- cls
			}
			wg.Done()
		}(ch, &wg)
	}

	// close the channel when all regions have finished the queries
	go func(wg *sync.WaitGroup, out chan<- *cluster.Cluster) {
		defer close(out)
		wg.Wait()
	}(&wg, ch)

	for c := range ch {
		// add DigitalOcean specific auth config
		c.GenerateAuthInfo = func(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
			return getConfigAuthInfo(cls)
		}
		clusters = append(clusters, c)
	}

	return clusters
}