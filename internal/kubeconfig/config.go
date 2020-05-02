// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClusterExporter interface {
	GetConfigCluster() *clientcmdapi.Cluster
	GetConfigAuthInfo() *clientcmdapi.AuthInfo
	GetUniqueId() string
}

// GetDefaultKubeconfigPath Returns the default path for the kubeconfig file
// based on the system
func GetDefaultKubeconfigPath() string {
	return clientcmd.RecommendedHomeFile
}

func getKubeConfig(kubeconfigPath string) (*clientcmdapi.Config, error) {
	cfg, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if cfg == nil {
		return clientcmdapi.NewConfig(), nil
	}
	return cfg, nil
}
