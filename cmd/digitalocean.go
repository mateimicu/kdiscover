// Package cmd offers CLI functionality
package cmd

import (
	"github.com/mateimicu/kdiscover/internal/digitalocean"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	doRegions []string
	doAlias   string
)

func newDigitalOceanCommand() *cobra.Command {
	DOCommand := &cobra.Command{
		Use:     "digitalocean",
		Aliases: []string{"do"},
		Short:   "Work with DigitalOcean DOKS clusters",
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
				"regions": doRegions,
			}).Debug("Using DigitalOcean regions")

			// Validate regions
			for _, region := range doRegions {
				if !digitalocean.IsValidRegion(region) {
					log.WithFields(log.Fields{
						"region": region,
					}).Warn("Unknown DigitalOcean region")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	DOCommand.PersistentFlags().StringSliceVar(
		&doRegions,
		"do-regions",
		digitalocean.GetRegions(),
		"Regions to search for DOKS clusters")
	DOCommand.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		kubeconfig.GetDefaultKubeconfigPath(),
		"Path to the kubeconfig to work with")
	DOCommand.PersistentFlags().StringVar(
		&doAlias,
		"cluster-name-template",
		"{{.Name}}",
		"Template to use for the cluster context name")

	DOCommand.AddCommand(newDOListCommand())
	DOCommand.AddCommand(newDOUpdateCommand())

	return DOCommand
}