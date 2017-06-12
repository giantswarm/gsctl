package commands

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
	"github.com/spf13/cobra"
)

const (
	listKeypairsActivityName string = "list-keypairs"
)

var (

	// ListKeypairsCommand performs the "list keypairs" function
	ListKeypairsCommand = &cobra.Command{
		Use:     "keypairs",
		Short:   "List key pairs for a cluster",
		Long:    `Prints a list of key pairs for a cluster`,
		PreRunE: checkListKeypairs,
		Run:     listKeypairs,
	}
)

func init() {
	ListKeypairsCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to list key pairs for")
	ListCommand.AddCommand(ListKeypairsCommand)
}

func checkListKeypairs(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster(requestIDHeader, listKeypairsActivityName, cmdLine, cmdAPIEndpoint)
		if clusterID != "" {
			cmdClusterID = clusterID
		} else {
			return errors.New("No cluster given. Please use the -c/--cluster flag to set a cluster ID.")
		}
	}
	return nil
}

// listKeypairs is the function called to list keypairs and display
// errors in case they happen
func listKeypairs(cmd *cobra.Command, args []string) {
	output, err := keypairsTable()
	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		if _, ok := err.(APIError); ok {
			dumpAPIResponse((err).(APIError).APIResponse)
		}
		os.Exit(1)
	}
	fmt.Print(output)
}

func keypairsTable() (string, error) {
	clientConfig := client.Configuration{
		Endpoint:  cmdAPIEndpoint,
		Timeout:   20 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient := client.NewClient(clientConfig)
	authHeader := "giantswarm " + config.Config.Token
	keypairsResponse, apiResponse, err := apiClient.GetKeyPairs(authHeader, cmdClusterID, requestIDHeader, listKeypairsActivityName, cmdLine)
	if err != nil {
		return "", APIError{err.Error(), *apiResponse}
	}

	if apiResponse.StatusCode == 200 {
		// sort result
		slice.Sort(keypairsResponse[:], func(i, j int) bool {
			return keypairsResponse[i].CreateDate < keypairsResponse[j].CreateDate
		})

		// create output
		output := []string{color.CyanString("CREATED") + "|" + color.CyanString("EXPIRES") + "|" + color.CyanString("ID") + "|" + color.CyanString("DESCRIPTION")}
		for _, keypair := range keypairsResponse {
			created := util.ShortDate(util.ParseDate(keypair.CreateDate))
			expires := util.ParseDate(keypair.CreateDate).Add(time.Duration(keypair.TtlHours) * time.Hour)

			// TODO: skip if expired
			output = append(output, created+"|"+
				util.ShortDate(expires)+"|"+
				util.Truncate(util.CleanKeypairID(keypair.Id), 10)+"|"+
				keypair.Description)
		}
		return columnize.SimpleFormat(output), nil
	}

	return "", APIError{
		fmt.Sprintf("Unhandled response code: %d", apiResponse.StatusCode),
		*apiResponse}
}
