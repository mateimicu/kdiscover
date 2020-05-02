// Package internal provides function to update kubeconfigs
package kubeconfig

import (
	"os/exec"
	"regexp"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

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
