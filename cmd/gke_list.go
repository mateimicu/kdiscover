// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	
	"github.com/mateimicu/kdiscover/internal/gcp"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newGKEListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all GKE Clusters",
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

			cmd.Println(getTable(convertToInterfaces(remoteGKEClusters), k, alias))
			return nil
		},
	}

	return listCommand
}

func getProjectZones() []gcp.ProjectZone {
	var projectZones []gcp.ProjectZone

	if len(gcpProjects) > 0 && len(gcpZones) > 0 {
		// Use explicit projects and zones
		for _, project := range gcpProjects {
			for _, zone := range gcpZones {
				projectZones = append(projectZones, gcp.ProjectZone{
					ProjectID: project,
					Zone:      zone,
				})
			}
		}
	} else if len(gcpProjects) > 0 {
		// Use explicit projects but discover zones
		for _, project := range gcpProjects {
			pzs := gcp.GetProjectsAndZones([]string{project})
			projectZones = append(projectZones, pzs...)
		}
	} else {
		// Auto-discover projects and zones
		projectZones = gcp.GetProjectsAndZones([]string{})
		if len(projectZones) == 0 {
			// Fall back to default if discovery fails
			projectZones = gcp.GetDefaultProjectsAndZones()
		}
	}

	log.WithFields(log.Fields{
		"project-zones-count": len(projectZones),
	}).Info("Using project/zone combinations")

	return projectZones
}