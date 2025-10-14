// Package azure provides wrapper for creating AKS cluster access
package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type AKSClient struct {
	Client         *armcontainerservice.ManagedClustersClient
	SubscriptionID string
	Location       string
}

func (c *AKSClient) String() string {
	return fmt.Sprintf("AKS Client for subscription %v location %v", c.SubscriptionID, c.Location)
}

func (c *AKSClient) GetClusters(ch chan<- *cluster.Cluster) {
	ctx := context.Background()
	
	// List all managed clusters in the subscription
	pager := c.Client.NewListPager(nil)
	
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"svc": c.String(),
			}).Warn("Can't list clusters")
			break
		}

		log.WithFields(log.Fields{
			"svc":    c.String(),
			"count":  len(page.Value),
		}).Debug("Parse page")

		for _, aksCluster := range page.Value {
			if aksCluster.Name == nil {
				continue
			}
			
			log.WithFields(log.Fields{
				"svc":     c.String(),
				"cluster": *aksCluster.Name,
			}).Debug("Found cluster")
			
			if cls, err := c.detailCluster(ctx, aksCluster); err == nil {
				ch <- cls
			} else {
				log.WithFields(log.Fields{
					"svc":     c.String(),
					"cluster": *aksCluster.Name,
					"err":     err,
				}).Warn("Can't get details on the cluster")
			}
		}
	}

	close(ch)
}

func (c *AKSClient) detailCluster(ctx context.Context, aksCluster *armcontainerservice.ManagedCluster) (*cluster.Cluster, error) {
	if aksCluster.Name == nil || aksCluster.ID == nil {
		return nil, errors.New("cluster name or ID is nil")
	}

	// Parse resource group from cluster ID
	// AKS cluster ID format: /subscriptions/{subscription-id}/resourceGroups/{resource-group}/providers/Microsoft.ContainerService/managedClusters/{cluster-name}
	resourceGroupName, err := extractResourceGroupFromID(*aksCluster.ID)
	if err != nil {
		return nil, fmt.Errorf("can't extract resource group from cluster ID %v: %w", *aksCluster.ID, err)
	}

	// Get cluster credentials
	response, err := c.Client.ListClusterAdminCredentials(ctx, resourceGroupName, *aksCluster.Name, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"cluster-name": *aksCluster.Name,
			"svc":          c.String(),
		}).Warn("Can't get cluster credentials")
		return nil, fmt.Errorf("can't fetch cluster credentials for %v: %w", *aksCluster.Name, err)
	}

	if len(response.Kubeconfigs) == 0 {
		return nil, fmt.Errorf("no kubeconfig found for cluster %v", *aksCluster.Name)
	}

	// Extract certificate authority data from the first kubeconfig
	kubeconfig := response.Kubeconfigs[0]
	var certificateAuthorityData string
	if kubeconfig.Value != nil {
		// Parse the kubeconfig to extract the certificate authority data
		certificateAuthorityData = string(kubeconfig.Value)
	}

	cls := cluster.NewCluster()
	cls.Provider = cluster.Azure
	cls.Name = *aksCluster.Name
	cls.ID = *aksCluster.ID
	
	if aksCluster.Properties != nil && aksCluster.Properties.Fqdn != nil {
		cls.Endpoint = fmt.Sprintf("https://%s", *aksCluster.Properties.Fqdn)
	}
	
	cls.CertificateAuthorityData = certificateAuthorityData
	
	if aksCluster.Properties != nil && aksCluster.Properties.ProvisioningState != nil {
		cls.Status = string(*aksCluster.Properties.ProvisioningState)
	}
	
	if aksCluster.Location != nil {
		cls.Region = *aksCluster.Location
	}
	
	// Set AKS-specific authentication
	cls.GenerateAuthInfo = func(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
		return getConfigAuthInfo(cls)
	}

	return cls, nil
}

// extractResourceGroupFromID extracts the resource group name from an Azure resource ID
func extractResourceGroupFromID(resourceID string) (string, error) {
	// Expected format: /subscriptions/{subscription-id}/resourceGroups/{resource-group}/providers/Microsoft.ContainerService/managedClusters/{cluster-name}
	parts := strings.Split(resourceID, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid resource ID format: %s", resourceID)
	}
	
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	
	return "", fmt.Errorf("resource group not found in resource ID: %s", resourceID)
}

func NewAKS(subscriptionID, location string) (*AKSClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.WithFields(log.Fields{
			"subscription": subscriptionID,
			"location":     location,
			"error":        err.Error(),
		}).Error("Failed to create Azure credential")
		return nil, err
	}

	client, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"subscription": subscriptionID,
			"location":     location,
			"error":        err.Error(),
		}).Error("Failed to create AKS client")
		return nil, err
	}

	return &AKSClient{
		Client:         client,
		SubscriptionID: subscriptionID,
		Location:       location,
	}, nil
}