// Package list holds the 'list *' sub-commands.
package list

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/list/clusters"
	"github.com/giantswarm/gsctl/commands/list/endpoints"
	"github.com/giantswarm/gsctl/commands/list/keypairs"
	"github.com/giantswarm/gsctl/commands/list/nodepools"
	"github.com/giantswarm/gsctl/commands/list/organizations"
	"github.com/giantswarm/gsctl/commands/list/releases"
)

var (
	// Command is the command to list things.
	Command = &cobra.Command{
		Use:   "list",
		Short: "List things, like organizations, clusters, key pairs",
		Long:  `Prints a list of the things you have access to.`,
	}
)

func init() {
	Command.AddCommand(clusters.Command)
	Command.AddCommand(endpoints.Command)
	Command.AddCommand(keypairs.Command)
	Command.AddCommand(nodepools.Command)
	Command.AddCommand(organizations.Command)
	Command.AddCommand(releases.Command)
}
