// Package internal provides function for working with EKS cluseters
package aws

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"os/exec"
	"regexp"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type EKSCluster cluster.Cluster

const (
	useAWSCLI = iota
	useIAMAuthenticator
)

const (
	commandAWScli           = "aws"
	commandIAMAuthenticator = "aws-iam-authenticator"
	clientAPIVersion        = "client.authentication.k8s.io/v1alpha1"
)

var (
	commands map[int]string = map[int]string{
		useAWSCLI:           commandAWScli,
		useIAMAuthenticator: commandIAMAuthenticator,
	}

	options map[int][]string = map[int][]string{
		useAWSCLI:           {"eks", "get-token", "--cluster-name"},
		useIAMAuthenticator: {"token", "-i"},
	}
)

func (cls *EKSCluster) GetConfigAuthInfo() *clientcmdapi.AuthInfo {
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

func getNewCluster(clsName string, svc *eks.EKS) (EKSCluster, error) {
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
		return EKSCluster{}, errors.New(msg)
	}
	certificatAuthorityData, err := base64.StdEncoding.DecodeString(*result.Cluster.CertificateAuthority.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"cluster-name":               *result.Cluster.Name,
			"arn":                        *result.Cluster.Arn,
			"certificate-authority-data": *result.Cluster.CertificateAuthority.Data,
		}).Error("Can't decode the Certificate Authority Data")
	}

	return EKSCluster{
		Name:                     *result.Cluster.Name,
		Id:                       *result.Cluster.Arn,
		Endpoint:                 *result.Cluster.Endpoint,
		CertificateAuthorityData: string(certificatAuthorityData),
		Status:                   *result.Cluster.Status,
	}, nil
}

func getEKSClustersPerRegion(region string, ch chan<- EKSCluster, wg *sync.WaitGroup) {
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
func GetEKSClusters(regions []string) []EKSCluster {
	clusters := make([]EKSCluster, 0, len(regions))

	var wg sync.WaitGroup
	ch := make(chan EKSCluster)

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
		case <-done:
			break Loop
		}
	}
	return clusters
}

func getAuthType() int {
	// According to the docs the first version that supports this is 1.18.17
	// See: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	// but looking at the source code the get token is present from 1.16.266
	// See: https://github.com/aws/aws-cli/commits/develop/awscli/customizations/eks/get_token.py
	pivotVersion, _ := semver.NewVersion("1.16.266")
	currentVersion := getAWSCLIversion()
	if currentVersion.LessThan(pivotVersion) {
		return useIAMAuthenticator
	}
	return useAWSCLI
}

func getAWSCLIversion() *semver.Version {
	v, _ := semver.NewVersion("0.0.0")
	command := exec.Command("aws", "--version")
	out, err := command.Output()
	if err != nil {
		log.Warn("Can't get aws cli tool version")
		return v
	}
	r := regexp.MustCompile(`aws-cli\/(?P<version>[0-9]+\.[0-9]+\.[0-9]+)`)
	if match := r.FindStringSubmatch(string(out)); len(match) != 0 {
		v, _ = semver.NewVersion(match[1])
		log.WithFields(log.Fields{
			"version": v,
		}).Info("Found AWS CLI version")
	}
	return v
}

func (cls *EKSCluster) GetUniqueId() string {
	return fmt.Sprintf("%v-%v-%v-%v", cls.Provider, cls.Id, cls.Region, cls.Name)
}

func (cls *EKSCluster) GetConfigCluster() *clientcmdapi.Cluster {
	cluster := clientcmdapi.NewCluster()
	cluster.Server = cls.Endpoint
	cluster.CertificateAuthorityData = []byte(cls.CertificateAuthorityData)
	return cluster
}

func (cls *EKSCluster) GetName() string {
	return cls.Name
}

func (cls *EKSCluster) GetRegion() string {
	return cls.Region
}

func (cls *EKSCluster) GetStatus() string {
	return cls.Status
}

func (cls *EKSCluster) GetEndpoint() string {
	return cls.Endpoint
}

func (cls EKSCluster) GetContextName(templateValue string) (string, error) {
	tmpl, err := template.New("context-name").Parse(templateValue)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, cls)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}
