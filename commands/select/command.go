package commands

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/commands/select/endpoint"
)

var (
	// Command is the command to list things
	Command = &cobra.Command{
		Use:   "select",
		Short: "Select things, like the API endpoint to use",
		Long:  `Select things, like the API endpoint to use`,
	}
)

func init() {
	Command.AddCommand(endpoint.Command)
}
