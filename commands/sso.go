package commands

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/pkce"
	"github.com/spf13/cobra"
)

var (
	// SSOCommand performs the "sso" function
	SSOCommand = &cobra.Command{
		Use:   "sso",
		Short: "Single Sign on for Admins",
		Long:  `Prints a list of all clusters you have access to`,
		Run:   ssoRunOutput,
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
	rand.Seed(time.Now().UTC().UnixNano())
	RootCommand.AddCommand(SSOCommand)
}

func ssoRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultSSOArguments()

	pkceResponse, err := pkce.Run()

	if pkceResponse.Error != "" {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println(pkceResponse.Error + ": " + pkceResponse.ErrorDescription)
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Try to parse the ID Token.
	idToken, err := pkce.ParseIdToken(pkceResponse.IDToken)
	if err != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to parse the ID Token.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Check if the access token works by fetching the installation's name.
	clientConfig := client.Configuration{
		Endpoint:  args.apiEndpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, err := client.NewClient(clientConfig)
	if err != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to verify token by fetching installation details.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	// Fetch installation name as alias.
	authHeader := "Bearer " + pkceResponse.AccessToken
	infoResponse, _, infoErr := apiClient.GetInfo(authHeader, requestIDHeader, loginActivityName, cmdLine)
	if infoErr != nil {
		fmt.Println(color.RedString("\nSomething went wrong during SSO."))
		fmt.Println("Unable to verify token by fetching installation details.")
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	alias := infoResponse.General.InstallationName

	// Store the token in the config file.
	if err := config.Config.StoreEndpointAuth(args.apiEndpoint, alias, idToken.Email, "Bearer", pkceResponse.AccessToken); err != nil {
		fmt.Println(color.RedString("\nSomething went while trying to store the token."))
		fmt.Println(err.Error())
		fmt.Println("Please notify the Giant Swarm support team, or try the command again in a few moments.")
		os.Exit(1)
	}

	fmt.Println(color.GreenString("\nYou are logged in as %s at %s.",
		idToken.Email, args.apiEndpoint))
}
