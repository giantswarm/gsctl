package commands

import "github.com/spf13/cobra"

var (
	// CreateCommand is the command to create things
	CreateCommand = &cobra.Command{
		Use:   "create",
		Short: "Create clusters, key pairs, ...",
		Long:  `Lets you create things like clusters, key pairs or kubectl configuration files`,
	}
)

const (
	// url to intallation instructions
	kubectlInstallURL = "http://kubernetes.io/docs/user-guide/prereqs/"

	// windows download page
	kubectlWindowsInstallURL = "https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md"
)

func init() {
	// subcommands
	RootCommand.AddCommand(CreateCommand)
}
