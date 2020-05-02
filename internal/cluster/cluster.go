// Package internal provides function for working with EKS cluseters
package cluster

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type K8sProvider int

const (
	None K8sProvider = iota
	AWS
	Google
	Azure
)

// Cluster is the representation of a K8S Cluster
// For now it is tailored to AWS, more specifically eks clusters
type Cluster struct {
	Provider                 K8sProvider
	Name                     string
	Region                   string
	Id                       string
	Endpoint                 string
	CertificateAuthorityData string
	Status                   string
}

func (cls *Cluster) GetConfigCluster() *clientcmdapi.Cluster {
	cluster := clientcmdapi.NewCluster()
	cluster.Server = cls.Endpoint
	cluster.CertificateAuthorityData = []byte(cls.CertificateAuthorityData)
	return cluster
}
