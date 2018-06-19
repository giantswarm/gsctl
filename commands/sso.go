package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

var (
	// SSOCommand performs the "sso" function
	SSOCommand = &cobra.Command{
		Use:    "sso",
		Short:  "Single Sign on for Admins",
		Long:   `Prints a list of all clusters you have access to`,
		PreRun: ssoPreRunOutput,
		Run:    ssoRunOutput,
	}
)

const (
	activityName = "sso"
)

type ssoArguments struct {
	apiEndpoint string
}

func defaultSSOArguments() ssoArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)

	return ssoArguments{
		apiEndpoint: endpoint,
	}
}

func init() {
	RootCommand.AddCommand(SSOCommand)
}

func ssoPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
}

func ssoRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultSSOArguments()
	fmt.Println(args)
	output := "test"

	if output != "" {
		fmt.Println(output)
	}
}
