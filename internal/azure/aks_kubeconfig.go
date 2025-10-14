// Package azure provides function for generating kubeconfig for AKS clusters
package azure

import (
	"github.com/mateimicu/kdiscover/internal/cluster"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientAPIVersion = "client.authentication.k8s.io/v1beta1"
)

func getConfigAuthInfo(cls *cluster.Cluster) *clientcmdapi.AuthInfo {
	authInfo := clientcmdapi.NewAuthInfo()
	
	// Use Azure CLI for authentication
	authInfo.Exec = &clientcmdapi.ExecConfig{
		APIVersion: clientAPIVersion,
		Command:    "kubelogin",
		Args: []string{
			"get-token",
			"--login", "azurecli",
			"--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630", // AKS AAD Server App ID
		},
		Env: []clientcmdapi.ExecEnvVar{
			{
				Name:  "AAD_SERVICE_PRINCIPAL_CLIENT_ID",
				Value: "",
			},
		},
		InteractiveMode: clientcmdapi.IfAvailableExecInteractiveMode,
	}
	
	return authInfo
}