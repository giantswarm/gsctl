// Package clusters implements the 'list clusters'  sub-command.
package clusters

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
	"github.com/giantswarm/gsctl/util"
)

var (
	// Command performs the "list clusters" function
	Command = &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "List clusters",
		Long:    `Prints a list of all clusters you have access to`,
		PreRun:  printValidation,
		Run:     printResult,
	}
)

const (
	listClustersActivityName = "list-clusters"
)

type Arguments struct {
	apiEndpoint       string
	authToken         string
	scheme            string
	userProvidedToken string
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		scheme:            scheme,
		userProvidedToken: flags.CmdToken,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()
	err := verifyListClusterPreconditions(args)

	if err == nil {
		return
	}

	errors.HandleCommonErrors(err)
}

func verifyListClusterPreconditions(args Arguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}

	return nil
}

// listClusters prints a table with all clusters the user has access to
func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	args := collectArguments()

	output, err := clustersTable(args)
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

	if output != "" {
		fmt.Println(output)
	}
}

// clustersTable returns a table of clusters the user has access to
func clustersTable(args Arguments) (string, error) {
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return "", microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listClustersActivityName

	response, err := clientWrapper.GetClusters(auxParams)
	if err != nil {
		if clienterror.IsUnauthorizedError(err) {
			return "", microerror.Mask(errors.NotAuthorizedError)
		}
		if clienterror.IsAccessForbiddenError(err) {
			return "", microerror.Mask(errors.AccessForbiddenError)
		}

		return "", microerror.Mask(err)
	}

	if len(response.Payload) == 0 {
		return color.YellowString("No clusters"), nil
	}
	// table headers
	output := []string{strings.Join([]string{
		color.CyanString("ID"),
		color.CyanString("ORGANIZATION"),
		color.CyanString("NAME"),
		color.CyanString("RELEASE"),
		color.CyanString("CREATED"),
	}, "|")}

	// sort clusters by ID
	sort.Slice(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	for _, cluster := range response.Payload {
		created := util.ShortDate(util.ParseDate(cluster.CreateDate))
		releaseVersion := cluster.ReleaseVersion
		if releaseVersion == "" {
			releaseVersion = "n/a"
		}

		output = append(output, strings.Join([]string{
			cluster.ID,
			cluster.Owner,
			cluster.Name,
			releaseVersion,
			created,
		}, "|"))
	}

	return columnize.SimpleFormat(output), nil
}
