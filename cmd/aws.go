// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"os"

	"github.com/mateimicu/kdiscover/internal"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	awsPartitions []string
	awsRegions    []string
)

func newAWSCommand() *cobra.Command {
	AWSCommand := &cobra.Command{
		Use:   "aws",
		Short: "Work with AWS EKS clusters",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.WithFields(log.Fields{
				"partitions": awsPartitions,
			}).Debug("Search regions for partitions")

			awsRegions = internal.GetRegions(awsPartitions)

			if len(awsRegions) == 0 {
				log.WithFields(log.Fields{
					"partitions": awsPartitions,
				}).Error("Can't find regions for partitions")
				os.Exit(errorExitCode)
			}

			log.WithFields(log.Fields{
				"regions": awsRegions,
			}).Info("Founds regions")
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	AWSCommand.PersistentFlags().StringSliceVar(
		&awsPartitions,
		"aws-partitions",
		[]string{"aws"},
		fmt.Sprintf("In what partitions to search for clusters. Supported %v", internal.AllowedParitions()))

	AWSCommand.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		internal.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")

	AWSCommand.AddCommand(newListCommand())
	AWSCommand.AddCommand(newUpdateCommand())

	return AWSCommand
}
