// Package gcp provides wrapper for creating GCP sessions and working with GKE clusters
package gcp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2/google"
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type GKEClient struct {
	Service   *container.Service
	ProjectID string
	Zone      string
}

func (c *GKEClient) String() string {
	return fmt.Sprintf("GKE Client for project %v, zone %v", c.ProjectID, c.Zone)
}

// GetClusters fetches all GKE clusters in the specified project and zone
func (c *GKEClient) GetClusters(ch chan<- *cluster.Cluster) {
	defer close(ch)

	log.WithFields(log.Fields{
		"project": c.ProjectID,
		"zone":    c.Zone,
	}).Info("Start fetching clusters")

	parent := fmt.Sprintf("projects/%s/locations/%s", c.ProjectID, c.Zone)
	resp, err := c.Service.Projects.Locations.Clusters.List(parent).Do()
	if err != nil {
		log.WithFields(log.Fields{
			"project": c.ProjectID,
			"zone":    c.Zone,
			"error":   err.Error(),
		}).Error("Failed to list clusters")
		return
	}

	log.WithFields(log.Fields{
		"project":        c.ProjectID,
		"zone":          c.Zone,
		"clusters-count": len(resp.Clusters),
	}).Info("Found clusters")

	for _, gkeCluster := range resp.Clusters {
		cls, err := c.convertGKECluster(gkeCluster)
		if err != nil {
			log.WithFields(log.Fields{
				"cluster-name": gkeCluster.Name,
				"project":      c.ProjectID,
				"zone":         c.Zone,
				"error":        err.Error(),
			}).Warn("Failed to convert cluster")
			continue
		}
		ch <- cls
	}
}

func (c *GKEClient) convertGKECluster(gkeCluster *container.Cluster) (*cluster.Cluster, error) {
	if gkeCluster.MasterAuth == nil || gkeCluster.MasterAuth.ClusterCaCertificate == "" {
		return nil, errors.New("cluster CA certificate not available")
	}

	// Decode the base64 encoded certificate
	certificateData, err := base64.StdEncoding.DecodeString(gkeCluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %v", err)
	}

	cls := cluster.NewCluster()
	cls.Provider = cluster.Google
	cls.Name = gkeCluster.Name
	cls.Region = c.Zone
	cls.ID = gkeCluster.SelfLink
	cls.Endpoint = fmt.Sprintf("https://%s", gkeCluster.Endpoint)
	cls.CertificateAuthorityData = string(certificateData)
	cls.Status = gkeCluster.Status
	cls.GenerateAuthInfo = func(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
		return getGKEAuthInfo(cls, c.ProjectID)
	}

	log.WithFields(log.Fields{
		"cluster-name": cls.Name,
		"project":      c.ProjectID,
		"zone":         c.Zone,
		"status":       cls.Status,
		"endpoint":     cls.Endpoint,
	}).Info("Successfully converted cluster")

	return cls, nil
}

func NewGKE(projectID, zone string) (*GKEClient, error) {
	ctx := context.Background()

	// Use default credentials from environment
	creds, err := google.FindDefaultCredentials(ctx, container.CloudPlatformScope)
	if err != nil {
		log.WithFields(log.Fields{
			"project": projectID,
			"zone":    zone,
			"error":   err.Error(),
		}).Error("Failed to find default credentials")
		return nil, err
	}

	service, err := container.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.WithFields(log.Fields{
			"project": projectID,
			"zone":    zone,
			"error":   err.Error(),
		}).Error("Failed to create container service")
		return nil, err
	}

	return &GKEClient{
		Service:   service,
		ProjectID: projectID,
		Zone:      zone,
	}, nil
}