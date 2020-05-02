// Package internal provides function for working with EKS cluseters
package cluster

import (
	"bytes"
	"fmt"
	"html/template"

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
	ID                       string
	Endpoint                 string
	CertificateAuthorityData string
	Status                   string
	GenerateClusterConfig    func(cls *Cluster) *clientcmdapi.Cluster
	GenerateAuthInfo         func(cls *Cluster) *clientcmdapi.AuthInfo
}

func NewCluster() *Cluster {
	return &Cluster{
		GenerateClusterConfig: defaultGenerateClusterConfig,
	}
}

func (cls *Cluster) GetUniqueID() string {
	return fmt.Sprintf("%v-%v-%v-%v", cls.Provider, cls.ID, cls.Region, cls.Name)
}

func defaultGenerateClusterConfig(cls *Cluster) *clientcmdapi.Cluster {
	cluster := clientcmdapi.NewCluster()
	cluster.Server = cls.Endpoint
	cluster.CertificateAuthorityData = []byte(cls.CertificateAuthorityData)
	return cluster
}

func (cls *Cluster) GetConfigAuthInfo() *clientcmdapi.AuthInfo {
	return cls.GenerateAuthInfo(cls)
}

func (cls *Cluster) GetConfigCluster() *clientcmdapi.Cluster {
	return cls.GenerateClusterConfig(cls)
}

func (cls *Cluster) GetName() string {
	return cls.Name
}

func (cls *Cluster) GetRegion() string {
	return cls.Region
}

func (cls *Cluster) GetStatus() string {
	return cls.Status
}

func (cls *Cluster) GetEndpoint() string {
	return cls.Endpoint
}

func (cls *Cluster) PrettyName(templateValue string) (string, error) {
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
