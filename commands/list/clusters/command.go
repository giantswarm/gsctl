// Package clusters implements the 'list clusters'  sub-command.
package clusters

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/client/clusters"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/clustercache"
	"github.com/giantswarm/gsctl/pkg/sortable"
	"github.com/giantswarm/gsctl/pkg/table"

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

	cmdSelector string

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
	Command.Flags().StringVarP(&cmdSelector, "selector", "l", "", "Label selector query to filter clusters on.")
}

type Arguments struct {
	apiEndpoint       string
	authToken         string
	outputFormat      string
	scheme            string
	selector          string
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
		selector:          cmdSelector,
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

	// Display error
	fmt.Println(color.RedString(err.Error()))

	os.Exit(1)
}

func verifyListClusterPreconditions(args Arguments) error {
	if args.apiEndpoint == "" {
		return microerror.Mask(errors.EndpointMissingError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(errors.NotLoggedInError)
	}
	if args.outputFormat != outputFormatJSON && args.outputFormat != outputFormatTable {
		return microerror.Maskf(errors.OutputFormatInvalidError, "Output format '%s' is unknown", args.outputFormat)
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
	var err error
	clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
	if err != nil {
		return "", microerror.Mask(err)
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listClustersActivityName

	var response *clusters.GetClustersOK

	if args.selector != "" {
		params := &models.V5ListClustersByLabelRequest{
			Labels: &args.selector,
		}
		response, err = clientWrapper.GetClustersByLabel(params, auxParams)
	} else {
		response, err = clientWrapper.GetClusters(auxParams)
	}

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
		// sort clusters by ID
		sort.Slice(response.Payload[:], func(i, j int) bool {
			return response.Payload[i].ID < response.Payload[j].ID
		})

		var clusters []*models.V4ClusterListItem
		{
			for _, cluster := range response.Payload {
				if cluster.DeleteDate != nil && !args.showDeleting {
					continue
				}

				clusters = append(clusters, cluster)
			}
		}

		outputBytes, err := json.MarshalIndent(clusters, outputJSONPrefix, outputJSONIndent)
		if err != nil {
			return "", microerror.Mask(err)
		}

		return string(outputBytes), nil
	}

	headers := []table.Column{
		table.Column{
			Name:        "id",
			DisplayName: "ID",
		},
		table.Column{
			Name:        "organization",
			DisplayName: "ORGANIZATION",
		},
		table.Column{
			Name:        "name",
			DisplayName: "NAME",
		},
		table.Column{
			Name:        "release",
			DisplayName: "RELEASE",
			Sortable: sortable.Sortable{
				SortType: sortable.Types.Semver,
			},
		},
		table.Column{
			Name:        "created",
			DisplayName: "CREATED",
			Sortable: sortable.Sortable{
				SortType: sortable.Types.Date,
			},
		},
	}

	if args.showDeleting {
		headers = append(headers, table.Column{
			Name:        "deleting-since",
			DisplayName: "DELETING SINCE",
		})
	}

	cTable := table.New()
	cTable.SetColumns(headers)

	numDeletedClusters := 0
	numOtherClusters := 0
	clusterIDs := make([]string, 0, len(response.Payload))

	rows := make([][]string, 0, len(response.Payload))
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
			clusterIDs = append(clusterIDs, cluster.ID)

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

		rows = append(rows, fields)
	}
	cTable.SetRows(rows)
	err = cTable.SortByColumnName("id", sortable.Directions.ASC)
	if err != nil {
		return "", microerror.Mask(err)
	}

	clustercache.CacheIDs(args.apiEndpoint, clusterIDs)

	// This function's output string.
	output := ""

	// Only show table when there is content.
	if len(rows) > 0 {
		output += cTable.String()
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
