package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	loginActivityName = "login"
)

var (
	// cmdPassword is the password given via command line flag
	cmdPassword string

	// email address passed as a positional argument
	cmdEmail string

	// cmdSSO is the bool that triggers login via SSO.
	cmdSSO bool

	// LoginCommand is the "login" CLI command
	LoginCommand = &cobra.Command{
		Use:   "login <email> [-e|--endpoint <endpoint>]",
		Short: "Sign in as a user",
		Long: `Sign in against an endpoint with email address and password.

This will select the given endpoint for subsequent commands.

The password has to be entered interactively or given as -p / --password flag.

The -e or --endpoint argument can be omitted if an endpoint is already selected.`,
		Example: "  gsctl login user@example.com --endpoint api.example.com",
		Run:     loginPickMethod,
	}
)

func init() {
	LoginCommand.Flags().StringVarP(&cmdPassword, "password", "p", "", "Password. If not given, will be prompted interactively.")
	LoginCommand.Flags().BoolVarP(&cmdSSO, "sso", "", false, "Authenticate using Single Sign On through our identity provider.")
	RootCommand.AddCommand(LoginCommand)
}

func loginPickMethod(cmd *cobra.Command, args []string) {
	if cmdSSO && cmdPassword != "" {
		fmt.Println(color.RedString("The --password argument has no effect when using --sso."))
		fmt.Println("Please execute the command as 'gsctl login --sso'. See 'gsctl login --help' for details.")
		os.Exit(1)
	}

	if cmdSSO && cmdToken != "" {
		fmt.Println(color.RedString("The --auth-token argument has no effect when using --sso."))
		fmt.Println("Please execute the command as 'gsctl login --sso'. See 'gsctl login --help' for details.")
		os.Exit(1)
	}

	if cmdSSO {
		ssoRunOutput(cmd, args)
		return
	}

	loginPreRunOutput(cmd, args)
	loginRunOutput(cmd, args)
}
