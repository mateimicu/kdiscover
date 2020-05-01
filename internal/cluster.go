// Package internal provides function for working with EKS cluseters
package internal

import "k8s.io/client-go/tools/clientcmd"

// Cluster is the representation of a K8S Cluster
// For now it is tailored to AWS, more specifically eks clusters
type Cluster struct {
	Name                     string
	Region                   string
	Id                       string
	Endpoint                 string
	CertificateAuthorityData string
	Status                   string
}

// IsExported will check if the cluster is already exporter
// in the kubeconfig file
// We consider a cluster "exported" if we have:
// * a `cluster` with the same Endpoint
// * a context for the cluster
func (cls *Cluster) IsExported(kubeconfigPath string) bool {
	cfg := clientcmd.GetConfigFromFileOrDie(kubeconfigPath)
	for _, ctx := range cfg.Contexts {
		if cluster, ok := cfg.Clusters[ctx.Cluster]; ok {
			if cluster.Server == cls.Endpoint {
				return true
			}
		}
	}
	return false
}
