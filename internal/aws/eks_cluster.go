// Package aws provides function for working with EKS cluseters
package aws

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientAPIVersion = "client.authentication.k8s.io/v1alpha1"
)

func getConfigAuthInfo(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
	authType := getAuthType()
	authInfo := clientcmdapi.NewAuthInfo()
	args := make([]string, len(options[authType]))
	copy(args, options[authType])
	args = append(args, cls.Name)
	args = append(args, "--region", cls.Region)

	authInfo.Exec = &clientcmdapi.ExecConfig{
		Command:    commands[authType],
		Args:       args,
		APIVersion: clientAPIVersion}
	return authInfo
}

func getNewCluster(clsName string, svc *eks.EKS) (*cluster.Cluster, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(clsName),
	}

	result, err := svc.DescribeCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Warn(aerr.Error())
		} else {
			log.Warn(err.Error())
		}
		msg := fmt.Sprintf("Can't fetch more details for the cluster %v", clsName)
		log.Warn(msg)
		return &cluster.Cluster{}, errors.New(msg)
	}
	certificatAuthorityData, err := base64.StdEncoding.DecodeString(*result.Cluster.CertificateAuthority.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"cluster-name":               *result.Cluster.Name,
			"arn":                        *result.Cluster.Arn,
			"certificate-authority-data": *result.Cluster.CertificateAuthority.Data,
		}).Error("Can't decode the Certificate Authority Data")
	}

	cls := cluster.NewCluster()
	cls.Name = *result.Cluster.Name
	cls.ID = *result.Cluster.Arn
	cls.Endpoint = *result.Cluster.Endpoint
	cls.CertificateAuthorityData = string(certificatAuthorityData)
	cls.Status = *result.Cluster.Status
	cls.GenerateAuthInfo = getConfigAuthInfo

	return cls, nil
}

func getEKSClustersPerRegion(region string, ch chan<- *cluster.Cluster, wg *sync.WaitGroup) {
	sess, err := getAWSSession(region)
	if err != nil {
		log.WithFields(log.Fields{
			"region": region,
			"error":  err.Error(),
		}).Error("Failed to create AWS SDK session")
	}
	svc := getEKSClient(sess)

	input := &eks.ListClustersInput{}

	err = svc.ListClustersPages(input,
		func(page *eks.ListClustersOutput, lastPage bool) bool {
			for _, cluster := range page.Clusters {
				log.WithFields(log.Fields{
					"region":  region,
					"cluster": cluster,
					"page":    page,
				}).Debug("Found cluster")
				if cls, ok := getNewCluster(*cluster, svc); ok == nil {
					cls.Region = region
					ch <- cls
				}
			}
			if lastPage {
				log.Debug("hit last page")
				return false
			}
			return true
		})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Can't list clusters")
	}
	wg.Done()
}

// GetEKSClusters will query the given regions and return a list of
// clusters accesable. It will use the default credential chain for AWS
// in order to figure out the context for the API calls
func GetEKSClusters(regions []string) []*cluster.Cluster {
	clusters := make([]*cluster.Cluster, 0, len(regions))

	var wg sync.WaitGroup
	ch := make(chan *cluster.Cluster)

	for _, region := range regions {
		log.WithFields(log.Fields{
			"region": region,
		}).Info("Query clusters")

		wg.Add(1)
		go getEKSClustersPerRegion(region, ch, &wg)
	}

	done := make(chan struct{})
	go func(done chan<- struct{}, wg *sync.WaitGroup) {
		wg.Wait()
		done <- struct{}{}
	}(done, &wg)

loop:
	for {
		select {
		case cluster := <-ch:
			clusters = append(clusters, cluster)
		case <-done:
			break loop
		}
	}
	return clusters
}
