// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"io/ioutil"
	"os"

	cluster "github.com/mateimicu/kdiscover/internal/cluster"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClusterExporter interface {
	GetConfigCluster() *clientcmdapi.Cluster
	GetConfigAuthInfo() *clientcmdapi.AuthInfo
	GetUniqueID() string
}

// GetDefaultKubeconfigPath Returns the default path for the kubeconfig file
// based on the system
func GetDefaultKubeconfigPath() string {
	return clientcmd.RecommendedHomeFile
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
	err = ioutil.WriteFile(path, output, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubeconfig) AddCluster(cls ClusterExporter, ctxName string) {
	key := cls.GetUniqueID()
	k.cfg.AuthInfos[key] = cls.GetConfigAuthInfo()
	k.cfg.Clusters[key] = cls.GetConfigCluster()
	k.cfg.Contexts[ctxName] = getConfigContext(key)
}

func getConfigContext(ctxName string) *clientcmdapi.Context {
	ctx := clientcmdapi.NewContext()
	ctx.Cluster = ctxName
	ctx.AuthInfo = ctxName
	return ctx
}

// Return all the clusters from a kubeconfig file
// the data
func (k *Kubeconfig) GetClusters() (map[string]cluster.Cluster, error) {
	clusters := make(map[string]cluster.Cluster)
	for name, c := range k.cfg.Clusters {
		cls := cluster.Cluster{
			Endpoint:                 c.Server,
			CertificateAuthorityData: string(c.CertificateAuthorityData),
		}

		clusters[name] = cls
	}
	return clusters, nil
}

type Endpointer interface {
	GetEndpoint() string
}

// IsExported will check if the cluster is already exporter
// in the kubeconfig file
// We consider a cluster "exported" if we have:
// * a `cluster` with the same Endpoint
// * a context for the cluster
func (k *Kubeconfig) IsExported(cls Endpointer) bool {
	for _, ctx := range k.cfg.Contexts {
		if cluster, ok := k.cfg.Clusters[ctx.Cluster]; ok {
			if cluster.Server == cls.GetEndpoint() {
				return true
			}
		}
	}
	return false
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
