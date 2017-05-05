package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/spf13/cobra"
)

var (
	// ListOrgsCommand performs the "list organizations" function
	ListOrgsCommand = &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs", "organisations"},
		Short:   "List organizations",
		Long:    `Prints a list of the organizations you are a member of`,
		PreRunE: checkListOrgs,
		Run:     listOrgs,
	}
)

const (
	listOrganizationsActivityName string = "list-organizations"
)

func init() {
	ListCommand.AddCommand(ListOrgsCommand)
}

func checkListOrgs(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in.\nUse '" + config.ProgramName + " login' to login or '--auth-token' to pass a valid auth token.")
	}
	return nil
}

// list organizations the user is member of
func listOrgs(cmd *cobra.Command, args []string) {
	output, err := orgsTable()
	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		if _, ok := err.(APIError); ok {
			dumpAPIResponse((err).(APIError).APIResponse)
		}
		os.Exit(1)
	}
	fmt.Print(output)
}

func orgsTable() (string, error) {
	clientConfig := client.Configuration{
		Endpoint:  cmdAPIEndpoint,
		Timeout:   5 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)

	// if token is set via flags, we unauthenticate using this token
	authHeader := "giantswarm " + config.Config.Token
	if cmdToken != "" {
		authHeader = "giantswarm " + cmdToken
	}

	orgsResponse, apiResponse, err := apiClient.GetUserOrganizations(authHeader, requestIDHeader, listOrganizationsActivityName, cmdLine)
	if err != nil {
		return "", APIError{err.Error(), *apiResponse}
	}

	if orgsResponse.StatusCode == apischema.STATUS_CODE_DATA {
		var output string
		var organizations = orgsResponse.Data
		if len(organizations) == 0 {
			output = color.YellowString("No organizations available\n")
		} else {
			sort.Strings(organizations)
			output = color.CyanString("ORGANIZATION") + "\n"
			for _, orgName := range organizations {
				output = output + orgName + "\n"
			}
		}
		return output, nil
	}
	return "", APIError{fmt.Sprintf("Unhandled response code: %v", orgsResponse.StatusCode), *apiResponse}
}
