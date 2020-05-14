// Package aws provides function for generatig kubeconfig for EKS clusters
package aws

import (
	"os/exec"
	"regexp"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

type AuthType int

const (
	useAWSCLI AuthType = iota
	useIAMAuthenticator
)

const (
	commandAWScli           = "aws"
	commandIAMAuthenticator = "aws-iam-authenticator"
)

var (
	commands map[AuthType]string = map[AuthType]string{
		useAWSCLI:           commandAWScli,
		useIAMAuthenticator: commandIAMAuthenticator,
	}

	awsCLIVersionCommand []string = []string{"aws", "--version"}
)

func getAuthType() AuthType {
	// According to the docs the first version that supports this is 1.18.17
	// See: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	// but looking at the source code the get token is present from 1.16.266
	// See: https://github.com/aws/aws-cli/commits/develop/awscli/customizations/eks/get_token.py
	pivotVersion, _ := semver.NewVersion("1.16.266")
	currentVersion := getAWSCLIversion(awsCLIVersionCommand)
	if currentVersion.LessThan(pivotVersion) {
		return useIAMAuthenticator
	}
	return useAWSCLI
}

func getAWSCLIversion(cmd []string) *semver.Version {
	v, _ := semver.NewVersion("0.0.0")
	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
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
