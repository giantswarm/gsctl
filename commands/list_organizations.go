package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/microerror"
)

var (
	// ListOrgsCommand performs the "list organizations" function
	ListOrgsCommand = &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs", "organisations"},
		Short:   "List organizations",
		Long:    `Prints a list of the organizations you are a member of`,
		PreRun:  listOrgsPreRunOutput,
		Run:     listOrgsRunOutput,
	}
)

const (
	listOrgsActivityName = "list-organizations"
)

type listOrgsArguments struct {
	apiEndpoint string
	authToken   string
	scheme      string
}

// defaultListOrgsArguments creates arguments based on command line flags and config
func defaultListOrgsArguments() listOrgsArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return listOrgsArguments{
		apiEndpoint: endpoint,
		authToken:   token,
		scheme:      scheme,
	}
}

func init() {
	ListCommand.AddCommand(ListOrgsCommand)
}

func listOrgsPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultListOrgsArguments()
	err := verifyListOrgsPreconditions(args)
	if err == nil {
		return
	}

	handleCommonErrors(err)
}

func verifyListOrgsPreconditions(args listOrgsArguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}
	return nil
}

// listOrgsRunOutput fetches a list organizations the user is member of
// and prints it in tabular form, or prints errors of they occur.
//
// TODO: Refactor so that this function calls the client, receives structured
// data which can be tested, and creates user-friendly output.
func listOrgsRunOutput(cmd *cobra.Command, args []string) {
	output, err := orgsTable()
	if err != nil {
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
func orgsTable() (string, error) {
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = listOrgsActivityName

	response, err := ClientV2.GetOrganizations(auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusUnauthorized {
				return "", microerror.Mask(notAuthorizedError)
			} else if clientErr.HTTPStatusCode == http.StatusForbidden {
				return "", microerror.Mask(accessForbiddenError)
			}
		}

		return "", microerror.Mask(err)
	}

	var output string
	if len(response.Payload) == 0 {
		output = color.YellowString("No organizations available\n")
	} else {
		// sort orgs by Id
		slice.Sort(response.Payload[:], func(i, j int) bool {
			return response.Payload[i].ID < response.Payload[j].ID
		})

		output = color.CyanString("ORGANIZATION") + "\n"
		for _, org := range response.Payload {
			output = output + org.ID + "\n"
		}
	}

	return output, nil
}
