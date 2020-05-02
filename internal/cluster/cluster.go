// Package internal provides function for working with EKS cluseters
package cluster

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
