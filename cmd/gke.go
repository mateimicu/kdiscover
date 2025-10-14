// Package cmd offers CLI functionality
package cmd

import (
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	gcpProjects []string
	gcpZones    []string
)

func newGKECommand() *cobra.Command {
	GKECommand := &cobra.Command{
		Use:   "gke",
		Short: "Work with GCP GKE clusters",
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
				"projects": gcpProjects,
				"zones":    gcpZones,
			}).Debug("Search clusters in projects and zones")

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	GKECommand.PersistentFlags().StringSliceVar(
		&gcpProjects,
		"gcp-projects",
		[]string{},
		"GCP project IDs to search for clusters. If empty, will try to discover all accessible projects")
	GKECommand.PersistentFlags().StringSliceVar(
		&gcpZones,
		"gcp-zones",
		[]string{},
		"GCP zones to search for clusters. If empty, will search all zones in the projects")
	GKECommand.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		kubeconfig.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")
	GKECommand.PersistentFlags().StringVar(
		&alias,
		"context-name-alias",
		"{{.Name}}",
		"Template for the context name. Has access to Cluster type")

	GKECommand.AddCommand(newGKEListCommand(), newGKEUpdateCommand())
	return GKECommand
}