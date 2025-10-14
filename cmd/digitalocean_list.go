// Package cmd offers CLI functionality
package cmd

import (
	"github.com/mateimicu/kdiscover/internal/digitalocean"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDOListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all DOKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			remoteDOKSClusters := digitalocean.GetDOKSClusters(doRegions)
			log.Info(remoteDOKSClusters)
			k, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			cmd.Println(getTable(convertToInterfaces(remoteDOKSClusters), k, doAlias))
			return nil
		},
	}

	return listCommand
}