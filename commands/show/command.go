// Package show implements the 'show' command.
package show

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands"
)

var (
	// ShowCommand is the command to display single items
	ShowCommand = &cobra.Command{
		Use:   "show",
		Short: "Show things, like clusters, releases",
		Long:  `Print details of a cluster or a release`,
	}
)

func init() {
	commands.RootCommand.AddCommand(ShowCommand)
}
