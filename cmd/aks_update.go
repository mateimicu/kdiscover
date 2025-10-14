// Package cmd offers CLI functionality  
package cmd

import (
	"fmt"

	"github.com/mateimicu/kdiscover/internal/azure"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newAKSUpdateCommand() *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all AKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Update all AKS Clusters")
			remoteAKSClusters := azure.GetAKSClusters(azureSubscriptions)
			cmd.Printf("Found %v clusters remote\n", len(remoteAKSClusters))
			
			if len(remoteAKSClusters) == 0 {
				cmd.Println("No AKS clusters found")
				return nil
			}

			// Backup existing kubeconfig
			cmd.Printf("Backup kubeconfig to %v.bak\n", kubeconfigPath)
			err := kubeconfig.BackupKubeconfig(kubeconfigPath)
			if err != nil {
				log.WithFields(log.Fields{
					"kubeconfig-path": kubeconfigPath,
					"err":             err.Error(),
				}).Warn("Can't backup kubeconfig")
			}

			k, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				log.WithFields(log.Fields{
					"kubeconfig-path": kubeconfigPath,
					"err":             err.Error(),
				}).Warn("Can't load kubeconfig, creating a new one")
				k = kubeconfig.New()
			}

			for _, cls := range remoteAKSClusters {
				ctxName, err := cls.PrettyName(aksAlias)
				if err != nil {
					log.WithFields(log.Fields{
						"err":          err.Error(),
						"cluster-name": cls.GetName(),
					}).Warn("Fallback on name")
					ctxName = cls.GetName()
				}
				log.WithFields(log.Fields{
					"cluster-name": cls.GetName(),
					"context-name": ctxName,
				}).Debug("Add cluster to kubeconfig")

				k.AddCluster(cls, ctxName)
			}

			err = k.Persist(kubeconfigPath)
			if err != nil {
				return fmt.Errorf("can't persist kubeconfig to %v: %w", kubeconfigPath, err)
			}

			cmd.Printf("Updated kubeconfig at %v\n", kubeconfigPath)
			return nil
		},
	}

	return updateCommand
}