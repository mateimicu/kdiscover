// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"os"

	cluster "github.com/mateimicu/kdiscover/internal/cluster"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

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

// UpdateKubeconfig will parse the given path as a valid kubeconfig file and
// will try to append all the clusters. It requires a name generator for cluster name
// generation
func UpdateKubeconfig(clusters []cluster.Cluster, kubeconfigPath string, gen ContextNameGenerator) error {
	kubeconfig, err := LoadKubeconfig(kubeconfigPath)
	//cfg, err := getKubeConfig(kubeconfigPath)
	if err != nil {
		return err
	}

	for _, cls := range clusters {
		ctxName, err := gen.GetContextName(cls)
		if err != nil {
			log.WithFields(log.Fields{
				"cluster": cls,
				"error":   err,
			}).Info("Can't generate alias for the cluster")
			continue
		}
		kubeconfig.AddCluster(cls, ctxName)
	}

	return nil
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

func getConfigCluster(cls cluster.Cluster) *clientcmdapi.Cluster {
	cluster := clientcmdapi.NewCluster()
	cluster.Server = cls.Endpoint
	cluster.CertificateAuthorityData = []byte(cls.CertificateAuthorityData)
	return cluster
}

func getConfigAuthInfo(cls cluster.Cluster, authType int) *clientcmdapi.AuthInfo {
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

func getConfigContext(name string) *clientcmdapi.Context {
	ctx := api.NewContext()
	ctx.Cluster = name
	ctx.AuthInfo = name
	return ctx
}
