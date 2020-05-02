// Package cmd offers CLI functionality
package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/mateimicu/kdiscover/internal/aws"
	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type clusterDescribe interface {
	GetEndpoint() string
	GetName() string
	GetRegion() string
	GetStatus() string
}

type exportable interface {
	IsExported(cls kubeconfig.Endpointer) bool
}

func getExportedString(e exportable, cls kubeconfig.Endpointer) string {
	if e.IsExported(cls) {
		return "Yes"
	}
	return "No"
}

func convertToInterfaces(clusters []*cluster.Cluster) []clusterDescribe {
	cls := make([]clusterDescribe, len(clusters))
	for i, c := range clusters {
		cls[i] = clusterDescribe(c)
	}
	return cls
}

func getTable(clusters []clusterDescribe, e exportable) string {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"Cluster Name", "Region", "Status", "Exported Locally"})
	rows := []table.Row{}
	for _, cls := range clusters {
		rows = append(rows, table.Row{cls.GetName(), cls.GetRegion(), cls.GetStatus(), getExportedString(e, cls)})
	}
	tw.AppendRows(rows)

	tw.AppendFooter(table.Row{"", "Number of clusters", len(clusters)})

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
	return tw.Render()
}

func newListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all EKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			remoteEKSClusters := aws.GetEKSClusters(awsRegions)
			log.Info(remoteEKSClusters)
			k, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			cmd.Println(getTable(convertToInterfaces(remoteEKSClusters), k))
			return nil
		},
	}

	return listCommand
}
