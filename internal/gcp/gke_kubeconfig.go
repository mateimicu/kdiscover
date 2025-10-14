// Package gcp provides function for generating kubeconfig for GKE clusters
package gcp

import (
	"github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientAPIVersion = "client.authentication.k8s.io/v1beta1"
)

func getGKEAuthInfo(cls *cluster.Cluster, projectID string) *clientcmdapi.AuthInfo {
	authInfo := clientcmdapi.NewAuthInfo()
	
	// Use gke-gcloud-auth-plugin for authentication
	authInfo.Exec = &clientcmdapi.ExecConfig{
		Command:            "gke-gcloud-auth-plugin",
		Args:               []string{},
		APIVersion:         clientAPIVersion,
		InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke",
		ProvideClusterInfo: true,
		InteractiveMode:    "Never",
		Env: []clientcmdapi.ExecEnvVar{
			{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "",
			},
		},
	}

	log.WithFields(log.Fields{
		"cluster-name": cls.Name,
		"project":      projectID,
	}).Debug("Generated GKE auth info")

	return authInfo
}