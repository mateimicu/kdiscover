// Package cmd offers CLI functionality
package cmd

import (
	"fmt"

	"github.com/mateimicu/kdiscover/internal/gcp"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newGKEUpdateCommand() *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all GKE Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectZones := getProjectZones()
			if len(projectZones) == 0 {
				log.Error("No project/zone combinations found")
				return fmt.Errorf("no project/zone combinations found")
			}

			remoteGKEClusters := gcp.GetGKEClusters(projectZones)
			log.Info(remoteGKEClusters)
			
			k, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			for _, cls := range remoteGKEClusters {
				prettyName, err := cls.PrettyName(alias)
				if err != nil {
					log.WithFields(log.Fields{
						"err":          err.Error(),
						"cluster-name": cls.GetName(),
					}).Warn("Fallback on name")
					prettyName = cls.GetName()
				}

				k.AddCluster(cls, prettyName)
				log.WithFields(log.Fields{
					"cluster-name": prettyName,
					"project":      cls.Region, // Note: We store project info in region field for consistency
				}).Info("Added cluster to kubeconfig")
			}

			err = k.Persist(kubeconfigPath)
			if err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"clusters-count": len(remoteGKEClusters),
				"kubeconfig":     kubeconfigPath,
			}).Info("Updated kubeconfig")

			return nil
		},
	}

	return updateCommand
}