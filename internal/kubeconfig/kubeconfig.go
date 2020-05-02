// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"io/ioutil"

	cluster "github.com/mateimicu/kdiscover/internal/cluster"
	"gopkg.in/yaml.v2"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ContextNameGenerator used for generating clusters names, the names
// are used for kubeconfig context
type ContextNameGenerator interface {
	GetContextName(cls cluster.Cluster) (string, error)
}

type Kubeconfig struct {
	cfg *clientcmdapi.Config
}

func LoadKubeconfig(kubeconfigpath string) (*Kubeconfig, error) {
	cfg, err := getKubeConfig(kubeconfigpath)
	if err != nil {
		return nil, err
	}

	return &Kubeconfig{
		cfg: cfg,
	}, nil
}

// Persist the kubeconfig to the disk
func (k *Kubeconfig) Persist(path string) error {
	output, err := yaml.Marshal(k.cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, output, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubeconfig) AddCluster(cls cluster.Cluster, ctxName string) {
	authType := getAuthType()
	key := cls.Id
	k.cfg.AuthInfos[key] = getConfigAuthInfo(cls, authType)
	k.cfg.Clusters[key] = getConfigCluster(cls)
	k.cfg.Contexts[ctxName] = getConfigContext(key)
}

// Return all the clusters from a kubeconfig file
// the data
func (k *Kubeconfig) GetClusters() (map[string]cluster.Cluster, error) {
	clusters := make(map[string]cluster.Cluster, 0)
	for name, c := range k.cfg.Clusters {
		cls := cluster.Cluster{
			Endpoint:                 c.Server,
			CertificateAuthorityData: string(c.CertificateAuthorityData),
		}

		clusters[name] = cls
	}
	return clusters, nil
}

// IsExported will check if the cluster is already exporter
// in the kubeconfig file
// We consider a cluster "exported" if we have:
// * a `cluster` with the same Endpoint
// * a context for the cluster
func (k *Kubeconfig) IsExported(cls *cluster.Cluster) bool {
	for _, ctx := range k.cfg.Contexts {
		if cluster, ok := k.cfg.Clusters[ctx.Cluster]; ok {
			if cluster.Server == cls.Endpoint {
				return true
			}
		}
	}
	return false
}
