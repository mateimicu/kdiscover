// Package cmd offers CLI functionality
package cmd

import (
	"fmt"

	"github.com/mateimicu/kdiscover/internal/azure"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	azureSubscriptions []string
	aksAlias           string
)

func newAKSCommand() *cobra.Command {
	AKSCommand := &cobra.Command{
		Use:   "aks",
		Short: "Work with Azure AKS clusters",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Because cobra will only run the last PersistentPreRunE
			// we need to search for the root and run the function.
			// This is not a robust solution as there may be multiple
			// PersistentPreRunE function between the leaf command and root
			// also this assumes that this is the root one
			// An issue about this https://github.com/spf13/cobra/issues/252
			root := cmd
			for ; root.HasParent(); root = root.Parent() {
			} //revive:disable-line:empty-block
			err := root.PersistentPreRunE(cmd, args)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"subscriptions": azureSubscriptions,
			}).Debug("Using Azure subscriptions")

			if len(azureSubscriptions) == 0 {
				log.WithFields(log.Fields{
					"subscriptions": azureSubscriptions,
				}).Error("No Azure subscriptions provided")
				return fmt.Errorf("no Azure subscriptions provided. Use --azure-subscriptions flag")
			}

			log.WithFields(log.Fields{
				"subscriptions": azureSubscriptions,
			}).Info("Found subscriptions")
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	AKSCommand.PersistentFlags().StringSliceVar(
		&azureSubscriptions,
		"azure-subscriptions",
		azure.GetDefaultSubscriptions(),
		"Azure subscription IDs to search for AKS clusters")
	AKSCommand.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		kubeconfig.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")
	AKSCommand.PersistentFlags().StringVar(
		&aksAlias,
		"context-name-alias",
		"{{.Name}}",
		"Template for the context name. Has access to Cluster type")

	AKSCommand.AddCommand(newAKSListCommand(), newAKSUpdateCommand())
	return AKSCommand
}