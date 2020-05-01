// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mateimicu/kdiscover/internal"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

const errorExitCode int = 1

var (
	kubeconfigPath string
	debug          bool
	rootCmd        = &cobra.Command{
		Use:   "kdiscover",
		Short: "Discover all EKS clusters on an account.",
		Long: `kdiscover is a simple utility that can search
all regions on an AWS account and try to find all EKS clsuters.
It will try to upgrade the kube-config for each cluster.`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetOutput(ioutil.Discard)
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
)

// Execute will create the tree of commands and will start parsing and execution
func Execute() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "set log level to debug")

	rootCmd.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		internal.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")

	rootCmd.AddCommand(newAWSCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(errorExitCode)
	}
}
