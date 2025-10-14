// Package cmd offers CLI functionality
package cmd

import (
	"github.com/mateimicu/kdiscover/internal/digitalocean"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDOUpdateCommand() *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all DOKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(cmd.Short)

			remoteDOKSClusters := digitalocean.GetDOKSClusters(doRegions)
			log.Info(remoteDOKSClusters)

			cmd.Printf("Found %v clusters remote\n", len(remoteDOKSClusters))

			if backupKubeconfig && fileExists(kubeconfigPath) {
				bName, err := backupKubeConfig(kubeconfigPath)
				if err != nil {
					return err
				}
				cmd.Printf("Backup kubeconfig to %v\n", bName)
			}
			kubeconfig, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			for _, cls := range remoteDOKSClusters {
				ctxName, err := cls.PrettyName(doAlias)
				if err != nil {
					log.WithFields(log.Fields{
						"cluster": cls,
						"error":   err,
					}).Info("Can't generate alias for the cluster")
					continue
				}
				kubeconfig.AddCluster(cls, ctxName)
			}
			err = kubeconfig.Persist(kubeconfigPath)
			if err != nil {
				cmd.Printf("Failed to persist kubeconfig %v", err.Error())
			}
			return err
		},
	}

	updateCommand.Flags().BoolVar(&backupKubeconfig, "backup-kubeconfig", true, "Backup cubeconfig before update")

	return updateCommand
}