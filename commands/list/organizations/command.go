// Package organizations implements the 'list organizations' sub-command.
package organizations

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

var (
	// Command performs the "list organizations" function
	Command = &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs", "organisations"},
		Short:   "List organizations",
		Long:    `Prints a list of the organizations you are a member of`,
		PreRun:  printValidation,
		Run:     printResult,
	}
)

const (
	listOrgsActivityName = "list-organizations"
)

type Arguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
}

// collectArguments creates arguments based on command line flags and config
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return Arguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()
	err := verifyListOrgsPreconditions(args)
	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
}

func verifyListOrgsPreconditions(args Arguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	return nil
}

// printResult fetches a list organizations the user is member of
// and prints it in tabular form, or prints errors of they occur.
//
// TODO: Refactor so that this function calls the client, receives structured
// data which can be tested, and creates user-friendly output.
func printResult(cmd *cobra.Command, extraArgs []string) {
	args := collectArguments()
	output, err := orgsTable(args)
	if err != nil {
		errors.HandleCommonErrors(err)
		client.HandleErrors(err)

		if clientErr, ok := err.(*clienterror.APIError); ok {
			fmt.Println(color.RedString(clientErr.ErrorMessage))
			if clientErr.ErrorDetails != "" {
				fmt.Println(clientErr.ErrorDetails)
			}
		} else {
			fmt.Println(color.RedString("Error: %s", err.Error()))
		}
		os.Exit(1)
	}

	fmt.Print(output)
}

// orgsTable fetches the organizations the user is a member of
// and returns a table in string form.
func orgsTable(args Arguments) (string, error) {
	clientV2, err := client.NewWithConfig(args.apiEndpoint, args.authToken)
	if err != nil {
		return "", microerror.Mask(err)
	}

	auxParams := clientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = listOrgsActivityName

	response, err := clientV2.GetOrganizations(auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusUnauthorized {
				return "", microerror.Mask(errors.NotAuthorizedError)
			} else if clientErr.HTTPStatusCode == http.StatusForbidden {
				return "", microerror.Mask(errors.AccessForbiddenError)
			}
		}

		return "", microerror.Mask(err)
	}

	var output string
	if len(response.Payload) == 0 {
		output = color.YellowString("No organizations available\n")
	} else {
		// sort orgs by Id
		sort.Slice(response.Payload[:], func(i, j int) bool {
			return response.Payload[i].ID < response.Payload[j].ID
		})

		output = color.CyanString("ORGANIZATION") + "\n"
		for _, org := range response.Payload {
			output = output + org.ID + "\n"
		}
	}

	return output, nil
}
