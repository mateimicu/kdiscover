// Package cmd offers CLI functionality
package cmd

import (
	"github.com/mateimicu/kdiscover/internal/azure"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newAKSListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all AKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			remoteAKSClusters := azure.GetAKSClusters(azureSubscriptions)
			log.Info(remoteAKSClusters)
			k, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			cmd.Println(getTable(convertToInterfaces(remoteAKSClusters), k, aksAlias))
			return nil
		},
	}

	return listCommand
}