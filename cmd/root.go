// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

const errorExitCode int = 1

var (
	loggingLevels map[string]log.Level = map[string]log.Level{
		"none":  0,
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
		"error": log.ErrorLevel,
		"warn":  log.WarnLevel,
		"info":  log.InfoLevel,
		"debug": log.DebugLevel,
		"trace": log.TraceLevel,
	}
	kubeconfigPath string
	logLevel       string
)

func NewRootCommand(version, commit, date, commandPrefix string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   commandPrefix,
		Short: "Discover all EKS clusters on an account.",
		Long: `kdiscover is a simple utility that can search
all regions on an AWS account and try to find all EKS clsuters.
It will try to upgrade the kube-config for each cluster.`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if logLevel == "none" {
				log.SetOutput(ioutil.Discard)
				return nil
			}

			if v, ok := loggingLevels[logLevel]; ok {
				log.SetLevel(v)
				return nil
			}

			return fmt.Errorf("can't find logging level %v", logLevel)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(
		&logLevel,
		"log-level",
		"none",
		fmt.Sprintf("Set logging lvl. Supported %v", getAllLogglingLevels()))

	rootCmd.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		kubeconfig.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")

	rootCmd.AddCommand(newAWSCommand())
	rootCmd.AddCommand(newVersionCommand(version, commit, date))
	return rootCmd
}

// Execute will create the tree of commands and will start parsing and execution
func Execute(version, commit, date string) {
	var prefix = "kdiscover"
	if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
		prefix = "kubectl discover"
	}
	rootCmd := NewRootCommand(version, commit, date, prefix)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(errorExitCode)
	}
}

func getAllLogglingLevels() []string {
	keys := make([]string, 0, len(loggingLevels))
	for k := range loggingLevels {
		keys = append(keys, k)
	}

	return keys
}
