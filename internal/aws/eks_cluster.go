// Package aws provides function for working with EKS cluseters
package aws

import (
	"sync"

	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientAPIVersion = "client.authentication.k8s.io/v1alpha1"
)

var (
	options map[AuthType][]string = map[AuthType][]string{
		useAWSCLI:           {"eks", "get-token", "--cluster-name"},
		useIAMAuthenticator: {"token", "-i"},
	}
)

func getConfigAuthInfo(cls *cluster.Cluster, authType AuthType) *clientcmdapi.AuthInfo {
	authInfo := clientcmdapi.NewAuthInfo()
	args := make([]string, len(options[authType]))
	copy(args, options[authType])
	args = append(args, cls.Name, "--region", cls.Region)

	authInfo.Exec = &clientcmdapi.ExecConfig{
		Command:    commands[authType],
		Args:       args,
		APIVersion: clientAPIVersion}
	return authInfo
}

type ClusterGetter interface {
	GetClusters(ch chan<- *cluster.Cluster)
}

func GetEKSClusters(regions []string) []*cluster.Cluster {
	clients := make([]ClusterGetter, 0, len(regions))

	for _, region := range regions {
		log.WithFields(log.Fields{
			"region": region,
		}).Info("Initialize client")
		eks, err := NewEKS(region)
		if err != nil {
			log.WithFields(log.Fields{
				"region": region,
				"error":  err.Error(),
			}).Error("Failed to create AWS SDK session")
			continue
		}

		clients = append(clients, ClusterGetter(eks))
	}
	return getEKSClusters(clients)
}

// GetEKSClusters will query the given regions and return a list of
// clusters accesable. It will use the default credential chain for AWS
// in order to figure out the context for the API calls
func getEKSClusters(clients []ClusterGetter) []*cluster.Cluster {
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

	// close the channel when all regions have finished the querys
	go func(wg *sync.WaitGroup, out chan<- *cluster.Cluster) {
		defer close(out)
		wg.Wait()
	}(&wg, ch)

	authType := getAuthType()
	for c := range ch {
		// add EKS specific auth config
		// create a new function in order to cache authType
		c.GenerateAuthInfo = func(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
			return getConfigAuthInfo(cls, authType)
		}
		clusters = append(clusters, c)
	}

	return clusters
}
