package internal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	log "github.com/sirupsen/logrus"
)

// Cluster is the representation of a K8S Cluster
// For now it is tailored to AWS, more specifically eks clusters
type Cluster struct {
	Name                     string
	Region                   string
	Arn                      string
	Endpoint                 string
	CertificateAuthorityData string
	Status                   string
}

func getNewCluster(clsName string, svc *eks.EKS) (Cluster, error) {
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
		return Cluster{}, errors.New(msg)
	}
	certificatAuthorityData, err := base64.StdEncoding.DecodeString(*result.Cluster.CertificateAuthority.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"cluster-name":               *result.Cluster.Name,
			"arn":                        *result.Cluster.Arn,
			"certificate-authority-data": *result.Cluster.CertificateAuthority.Data,
		}).Error("Can't decode the Certificate Authority Data")
	}

	return Cluster{
		Name:                     *result.Cluster.Name,
		Arn:                      *result.Cluster.Arn,
		Endpoint:                 *result.Cluster.Endpoint,
		CertificateAuthorityData: string(certificatAuthorityData),
		Status:                   *result.Cluster.Status,
	}, nil
}

func getEKSClustersPerRegion(region string, ch chan<- Cluster, wg *sync.WaitGroup) {
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
func GetEKSClusters(regions []string) []Cluster {
	var clusters []Cluster = make([]Cluster, 0, len(regions))

	var wg sync.WaitGroup
	ch := make(chan Cluster, 0)

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

Loop:
	for {
		select {
		case cluster := <-ch:
			clusters = append(clusters, cluster)
		case _ = <-done:
			break Loop
		}
	}
	return clusters
}
