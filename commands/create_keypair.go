package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
	"github.com/spf13/cobra"
)

var (
	// CreateKeypairCommand performs the "create keypair" function
	CreateKeypairCommand = &cobra.Command{
		Use:     "keypair",
		Short:   "Create key-pair",
		Long:    `Creates a new key-pair for a cluster`,
		PreRunE: checkAddKeypair,
		Run:     addKeypair,
	}
)

const (
	addKeyPairActivityName string = "add-keypair"
)

func init() {
	CreateKeypairCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key-pair for")
	CreateKeypairCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")

	CreateCommand.AddCommand(CreateKeypairCommand)
}

func checkAddKeypair(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster(requestIDHeader, addKeyPairActivityName, cmdLine, cmdAPIEndpoint)
		if clusterID != "" {
			cmdClusterID = clusterID
		} else {
			return errors.New("No cluster given. Please use the -c/--cluster flag to set a cluster ID.")
		}
	}
	if cmdDescription == "" {
		return errors.New("No description given. Please use the -d/--description flag to set a description.")
	}
	return nil
}

func addKeypair(cmd *cobra.Command, args []string) {
	if cmdDescription == "" {
		cmdDescription = "Added by user " + config.Config.Email + " using 'gsctl create keypair'"
	}

	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	authHeader := "giantswarm " + config.Config.Token
	ttlHours := int32(cmdTTLDays * 24)
	addKeyPairBody := gsclientgen.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}
	keypairResponse, apiResponse, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody, requestIDHeader, addKeyPairActivityName, cmdLine)

	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}

	if keypairResponse.StatusCode == apischema.STATUS_CODE_DATA {
		cleanID := util.CleanKeypairID(keypairResponse.Data.Id)
		msg := fmt.Sprintf("New key-pair created with ID %s", cleanID)
		fmt.Println(color.GreenString(msg))

		// store credentials to file
		caCertPath := util.StoreCaCertificate(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.CertificateAuthorityData)
		fmt.Println("CA certificate stored in:", caCertPath)

		clientCertPath := util.StoreClientCertificate(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientCertificateData)
		fmt.Println("Client certificate stored in:", clientCertPath)

		clientKeyPath := util.StoreClientKey(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientKeyData)
		fmt.Println("Client private key stored in:", clientKeyPath)

	} else {
		fmt.Println(color.RedString("Unhandled response code: %v", keypairResponse.StatusCode))
		dumpAPIResponse(*apiResponse)
	}
}
