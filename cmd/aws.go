// Package cmd offers CLI functionality
package cmd

import (
	"fmt"

	"github.com/mateimicu/kdiscover/internal/aws"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Because cobra will only run the last PersistentPreRunE
			// we need to search for the root and run the function.
			// This is not a robust solution as there may be multiple
			// PersistentPreRunE function between the leaf command and root
			// also this assumes that this is the root one
			// An issue about this https://github.com/spf13/cobra/issues/252
			root := cmd
			for ; root.HasParent(); root = root.Parent() {
			}
			err := root.PersistentPreRunE(cmd, args)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"partitions": awsPartitions,
			}).Debug("Search regions for partitions")

			awsRegions = aws.GetRegions(awsPartitions)

			if len(awsRegions) == 0 {
				log.WithFields(log.Fields{
					"partitions": awsPartitions,
				}).Error("Can't find regions for partitions")
				return fmt.Errorf("Can't find regions for partitions %v", awsPartitions)
			}

			log.WithFields(log.Fields{
				"regions": awsRegions,
			}).Info("Founds regions")
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	AWSCommand.PersistentFlags().StringSliceVar(
		&awsPartitions,
		"aws-partitions",
		[]string{"aws"},
		fmt.Sprintf("In what partitions to search for clusters. Supported %v", aws.AllowedParitions()))

	AWSCommand.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		kubeconfig.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")

	AWSCommand.AddCommand(newListCommand(), newUpdateCommand())
	return AWSCommand
}
