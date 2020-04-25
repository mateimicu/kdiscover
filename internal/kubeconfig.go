// Package internal provides function to update kubeconfigs
package internal

import (
	"os/exec"
	"regexp"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ContextNameGenerator used for generating clusters names, the names
// are used for kubeconfig context
type ContextNameGenerator interface {
	GetContextName(cls Cluster) (string, error)
}

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

func getConfigCluster(cls Cluster) *api.Cluster {
	cluster := api.NewCluster()
	cluster.Server = cls.Endpoint
	cluster.CertificateAuthorityData = []byte(cls.CertificateAuthorityData)
	return cluster
}

func getConfigAuthInfo(cls Cluster, authType int) *api.AuthInfo {
	authInfo := api.NewAuthInfo()
	args := make([]string, len(options[authType]))
	copy(args, options[authType])
	args = append(args, cls.Name)
	args = append(args, "--region", cls.Region)

	authInfo.Exec = &api.ExecConfig{
		Command:    commands[authType],
		Args:       args,
		APIVersion: clientAPIVersion}
	return authInfo
}

func getConfigContext(name string) *api.Context {
	ctx := api.NewContext()
	ctx.Cluster = name
	ctx.AuthInfo = name
	return ctx
}

func getAuthType() int {
	// According to the docs the first version that supports this is 1.18.17
	// See: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	// but looking at the source code the get token is present from 1.16.266
	// See: https://github.com/aws/aws-cli/commits/develop/awscli/customizations/eks/get_token.py
	pivotVersion, _ := semver.NewVersion("1.16.266")
	currentVersion := getAWSCLIversion()
	if currentVersion.LessThan(pivotVersion) {
		return useIAMAuthenticator
	}
	return useAWSCLI
}

// UpdateKubeconfig will parse the given path as a valid kubeconfig file and
// will try to append all the clusters. It requires a name generator for cluster name
// generation
func UpdateKubeconfig(clusters []Cluster, kubeconfigPath string, gen ContextNameGenerator) error {
	authType := getAuthType()
	cfg := clientcmd.GetConfigFromFileOrDie(kubeconfigPath)

	for _, cls := range clusters {
		key := cls.Arn
		cfg.AuthInfos[key] = getConfigAuthInfo(cls, authType)
		cfg.Clusters[key] = getConfigCluster(cls)
		ctxName, err := gen.GetContextName(cls)
		if err != nil {
			log.WithFields(log.Fields{
				"cluster": cls,
				"error":   err,
			}).Info("Can't generate alias for the cluster")
			continue
		}
		cfg.Contexts[ctxName] = getConfigContext(key)
	}

	err := clientcmd.WriteToFile(*cfg, kubeconfigPath)
	return err
}

func getAWSCLIversion() *semver.Version {
	v, _ := semver.NewVersion("0.0.0")
	command := exec.Command("aws", "--version")
	out, err := command.Output()
	if err != nil {
		log.Warn("Can't get aws cli tool version")
		return v
	}
	r := regexp.MustCompile(`aws-cli\/(?P<version>[0-9]+\.[0-9]+\.[0-9]+)`)
	if match := r.FindStringSubmatch(string(out)); len(match) != 0 {
		v, _ = semver.NewVersion(match[1])
		log.WithFields(log.Fields{
			"version": v,
		}).Info("Found AWS CLI version")
	}
	return v
}

// GetDefaultKubeconfigPath Returns the default path for the kubeconfig file
// based on the system
func GetDefaultKubeconfigPath() string {
	return clientcmd.RecommendedHomeFile
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
