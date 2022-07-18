// Package cmd offers CLI functionality
package cmd

import (
	"github.com/spf13/cobra"
	goversion "go.hein.dev/go-version"
)

var (
	shortened = false
	output    = "json"
)

func newVersionCommand(version, commit, date string) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  "",

		RunE: func(cmd *cobra.Command, args []string) error {
			resp := goversion.FuncWithOutput(shortened, version, commit, date, output)
			cmd.Print(resp)
			return nil
		},
	}

	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")

	return versionCmd
}
