// Package clusters implements the 'list clusters'  sub-command.
package clusters

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

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
		Long:    `Prints a list of all clusters you have access to.`,
		PreRun:  printValidation,
		Run:     printResult,
	}

	cmdOutput string

	cmdShowDeleted bool

	arguments Arguments
)

const (
	listClustersActivityName = "list-clusters"

	outputFormatJSON  = "json"
	outputFormatTable = "table"

	outputJSONPrefix = ""
	outputJSONIndent = "  "
)

func init() {
	initFlags()
}

func initFlags() {
	Command.ResetFlags()
	Command.Flags().StringVarP(&cmdOutput, "output", "o", "table", "Use 'json' for JSON output. Defaults to human-friendly table output.")
	Command.Flags().BoolVarP(&cmdShowDeleted, "show-deleting", "", false, "Show clusters which are currently being deleted (only with cluster release > 10.0.0).")
}

type Arguments struct {
	apiEndpoint       string
	authToken         string
	outputFormat      string
	scheme            string
	showDeleting      bool
	userProvidedToken string
}

func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		authToken:         token,
		outputFormat:      cmdOutput,
		scheme:            scheme,
		showDeleting:      cmdShowDeleted,
		userProvidedToken: flags.Token,
	}
}

func printValidation(cmd *cobra.Command, cmdLineArgs []string) {
	arguments = collectArguments()
	err := verifyListClusterPreconditions(arguments)

	if err == nil {
		return
	}

	client.HandleErrors(err)
	errors.HandleCommonErrors(err)

	if errors.IsConflictingFlagsError(err) {
		fmt.Println(color.RedString("Conflicting flags used"))
		fmt.Println("The --show-deleting flag cannot be used with JSON output.")
		fmt.Println("JSON output contains all clusters, including deleted, by default.")
	} else {
		fmt.Println(color.RedString(err.Error()))
	}

	os.Exit(1)
}

func verifyListClusterPreconditions(args Arguments) error {
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if args.outputFormat != outputFormatJSON && args.outputFormat != outputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, "Output format '%s' is unknown", args.outputFormat)
	}
	if args.outputFormat == outputFormatJSON && args.showDeleting == true {
		return microerror.Maskf(errors.ConflictingFlagsError, "The --show-deleting flag cannot be used with JSON output.")
	}

	return nil
}

// printResult prints a table with all clusters the user has access to
func printResult(cmd *cobra.Command, cmdLineArgs []string) {
	output, err := getClustersOutput(arguments)
	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

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

// getClustersOutput returns a table of clusters the user has access to
func getClustersOutput(args Arguments) (string, error) {
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

	if args.outputFormat == "json" {
		outputBytes, err := json.MarshalIndent(response.Payload, outputJSONPrefix, outputJSONIndent)
		if err != nil {
			return "", microerror.Mask(err)
		}

		return string(outputBytes), nil
	}

	// sort clusters by ID
	sort.Slice(response.Payload[:], func(i, j int) bool {
		return response.Payload[i].ID < response.Payload[j].ID
	})

	headers := []string{
		color.CyanString("ID"),
		color.CyanString("ORGANIZATION"),
		color.CyanString("NAME"),
		color.CyanString("RELEASE"),
		color.CyanString("CREATED"),
	}

	if args.showDeleting {
		headers = append(headers, color.CyanString("DELETING SINCE"))
	}

	table := []string{strings.Join(headers, "|")}

	numDeletedClusters := 0
	numOtherClusters := 0

	for _, cluster := range response.Payload {
		created := util.ShortDate(util.ParseDate(cluster.CreateDate))
		deleted := "n/a"

		var secondsSinceDelete float64

		if cluster.DeleteDate != nil {
			numDeletedClusters++

			if !args.showDeleting {
				continue
			}

			deleted = util.ShortDate(util.ParseDate(cluster.DeleteDate.String()))
			deleteTime := time.Time(*cluster.DeleteDate)
			secondsSinceDelete = time.Now().Sub(deleteTime).Seconds()
		} else {
			numOtherClusters++
		}

		releaseVersion := cluster.ReleaseVersion
		if releaseVersion == "" {
			releaseVersion = "n/a"
		}

		fields := []string{
			cluster.ID,
			cluster.Owner,
			cluster.Name,
			releaseVersion,
			created,
		}
		if args.showDeleting {
			fields = append(fields, color.RedString(deleted))
		}

		// Highlight row in red if old.
		if secondsSinceDelete > 86400 {
			for index := range fields {
				fields[index] = color.RedString(fields[index])
			}
		}

		table = append(table, strings.Join(fields, "|"))
	}

	// This function's output string.
	output := ""

	// Only show table when there is content.
	if len(table) > 1 {
		output += columnize.SimpleFormat(table)
	} else {
		output += color.YellowString("No clusters")
	}

	if !args.showDeleting && numDeletedClusters > 0 {
		output += "\n\n"
		if numDeletedClusters == 1 {
			output += fmt.Sprintf("There is 1 additional cluster currently being deleted. Add the %s flag to see it.", color.CyanString("--show-deleting"))
		} else {
			output += fmt.Sprintf("There are %d additional clusters currently being deleted. Add the %s flag to see them.", numDeletedClusters, color.CyanString("--show-deleting"))
		}
	}

	return output, nil
}
