// Package digitalocean provides wrapper for creating DigitalOcean sessions
package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type DOKSClient struct {
	Client *godo.Client
	Region string
}

func (c *DOKSClient) String() string {
	return fmt.Sprintf("DOKS Client for region %v", c.Region)
}

// GetClusters retrieves all Kubernetes clusters from DigitalOcean
func (c *DOKSClient) GetClusters(ch chan<- *cluster.Cluster) {
	ctx := context.Background()
	
	// List all clusters (DigitalOcean doesn't filter by region in list call)
	clusters, _, err := c.Client.Kubernetes.List(ctx, &godo.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"svc": c.String(),
		}).Warn("Can't list clusters")
		close(ch)
		return
	}

	log.WithFields(log.Fields{
		"svc":           c.String(),
		"cluster_count": len(clusters),
	}).Debug("Found clusters")

	for _, doCluster := range clusters {
		// Filter by region if specified
		if c.Region != "" && doCluster.RegionSlug != c.Region {
			continue
		}

		log.WithFields(log.Fields{
			"svc":        c.String(),
			"cluster_id": doCluster.ID,
			"name":       doCluster.Name,
			"region":     doCluster.RegionSlug,
		}).Debug("Processing cluster")

		if cls, err := c.convertCluster(doCluster); err == nil {
			ch <- cls
		} else {
			log.WithFields(log.Fields{
				"svc":        c.String(),
				"cluster_id": doCluster.ID,
				"err":        err,
			}).Warn("Can't convert cluster")
		}
	}

	close(ch)
}

func (c *DOKSClient) convertCluster(doCluster *godo.KubernetesCluster) (*cluster.Cluster, error) {
	if doCluster.Status == nil {
		return nil, errors.New("cluster status is nil")
	}

	// Create new cluster
	cls := cluster.NewCluster()
	cls.Provider = cluster.DigitalOcean
	cls.Name = doCluster.Name
	cls.ID = doCluster.ID
	cls.Region = doCluster.RegionSlug
	cls.Endpoint = doCluster.Endpoint
	cls.Status = string(doCluster.Status.State)

	// For DigitalOcean, we'll get the certificate authority data when generating kubeconfig
	// as it's not directly available in the cluster object
	cls.CertificateAuthorityData = ""

	return cls, nil
}

// NewDOKS creates a new DigitalOcean Kubernetes client
func NewDOKS(region string) (*DOKSClient, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, errors.New("DIGITALOCEAN_TOKEN environment variable is required")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &DOKSClient{
		Client: client,
		Region: region,
	}, nil
}