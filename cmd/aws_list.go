// Package cmd offers CLI functionality
package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/mateimicu/kdiscover/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all EKS Clusters",
		Run: func(cmd *cobra.Command, args []string) {
			remoteEKSClusters := internal.GetEKSClusters(awsRegions)
			log.Info(remoteEKSClusters)

			tw := table.NewWriter()
			tw.AppendHeader(table.Row{"Cluster Name", "Region", "Status", "Exported Locally"})
			rows := []table.Row{}
			for _, cls := range remoteEKSClusters {
				rows = append(rows, table.Row{cls.Name, cls.Region, cls.Status, getExportedString(cls, kubeconfigPath)})
			}
			tw.AppendRows(rows)

			tw.AppendFooter(table.Row{"", "Number of clusters", len(remoteEKSClusters)})

			tw.SetAutoIndex(true)
			tw.SortBy([]table.SortBy{{Name: "Region", Mode: table.Dsc}})

			tw.SetStyle(table.StyleLight)
			tw.Style().Format.Header = text.FormatLower
			tw.Style().Format.Footer = text.FormatLower
			tw.Style().Options.SeparateColumns = false
			tw.SetColumnConfigs([]table.ColumnConfig{
				{
					Name:        "Exported Locally",
					Align:       text.AlignCenter,
					AlignHeader: text.AlignCenter,
				},
			})
			// render it
			fmt.Println(tw.Render())
		},
	}

	return listCommand
}

func getExportedString(cls internal.Cluster, kubeconfigPath string) string {
	if cls.IsExported(kubeconfigPath) {
		return "Yes"
	}
	return "No"
}
