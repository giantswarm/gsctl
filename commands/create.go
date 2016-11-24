package commands

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// CreateCommand is the command to create things
	CreateCommand = &cobra.Command{
		Use:   "create",
		Short: "Create things, like kubectl configuration, or key-pairs",
		Long:  `Lets you create things like key-pairs`,
	}

	// CreateKeypairCommand performs the "create keypair" function
	CreateKeypairCommand = &cobra.Command{
		Use:     "keypair",
		Short:   "Create key-pair",
		Long:    `Creates a new key-pair for a cluster`,
		PreRunE: checkAddKeypair,
		Run:     addKeypair,
	}

	// CreateKubeconfigCommand performs the "create kubeconfig" function
	CreateKubeconfigCommand = &cobra.Command{
		Use:     "kubeconfig",
		Short:   "Configure kubectl",
		Long:    `Modifies kubectl configuration to access your Giant Swarm Kubernetes cluster`,
		PreRunE: checkCreateKubeconfig,
		Run:     createKubeconfig,
	}
)

func init() {
	CreateKeypairCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key-pair for")
	CreateKeypairCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")

	CreateKubeconfigCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")
	CreateKubeconfigCommand.Flags().IntVarP(&cmdTTLDays, "ttl", "", 30, "Duration until expiry of the created key-pair in days")

	// subcommands
	CreateCommand.AddCommand(CreateKeypairCommand, CreateKubeconfigCommand)
}

func checkAddKeypair(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster()
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
	client := gsclientgen.NewDefaultApi()
	authHeader := "giantswarm " + config.Config.Token
	ttlHours := int32(cmdTTLDays * 24)
	addKeyPairBody := gsclientgen.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}
	keypairResponse, _, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody)

	if err != nil {
		log.Fatal(err)
	}

	if keypairResponse.StatusCode == 10000 {
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
		fmt.Printf("Unhandled response code: %v", keypairResponse.StatusCode)
		fmt.Printf("Status text: %v", keypairResponse.StatusText)
	}
}

// Pre-check before creating a new kubeconfig
func checkCreateKubeconfig(cmd *cobra.Command, args []string) error {
	util.CheckKubectl()
	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster()
		if clusterID != "" {
			cmdClusterID = clusterID
		} else {
			return errors.New("No cluster given. Please use the -c/--cluster flag to set a cluster ID.")
		}
	}
	return nil
}

// createKubeconfig adds configuration for kubectl
func createKubeconfig(cmd *cobra.Command, args []string) {
	client := gsclientgen.NewDefaultApi()
	authHeader := "giantswarm " + config.Config.Token

	// parameters given by the user
	ttlHours := int32(cmdTTLDays * 24)
	if cmdDescription == "" {
		cmdDescription = "Added by user " + config.Config.Email + " using 'g8m create kubeconfig'"
	}

	addKeyPairBody := gsclientgen.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}

	fmt.Println("Creating new key-pair…")

	keypairResponse, _, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody)

	if err != nil {
		fmt.Println(color.RedString("Error in createKubeconfig:"))
		log.Fatal(err)
		fmt.Println("keypairResponse:", keypairResponse)
		fmt.Println("addKeyPairBody:", addKeyPairBody)
		os.Exit(1)
	}

	if keypairResponse.StatusCode == 10000 {
		msg := fmt.Sprintf("New key-pair created with ID %s and expiry of %v hours",
			util.Truncate(util.CleanKeypairID(keypairResponse.Data.Id), 10),
			ttlHours)
		fmt.Println(msg)

		// store credentials to file
		caCertPath := util.StoreCaCertificate(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.CertificateAuthorityData)

		clientCertPath := util.StoreClientCertificate(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientCertificateData)

		clientKeyPath := util.StoreClientKey(config.ConfigDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientKeyData)

		fmt.Println("Certificate and key files written to:")
		fmt.Println(caCertPath)
		fmt.Println(clientCertPath)
		fmt.Println(clientKeyPath)

		// TODO: Take this from the cluster object
		apiEndpoint := "https://api." + cmdClusterID + ".k8s.gigantic.io"

		// edit kubectl config
		util.KubectlSetCluster(cmdClusterID, apiEndpoint, caCertPath)
		util.KubectlSetCredentials(cmdClusterID, clientKeyPath, clientCertPath)
		util.KubectlSetContext(cmdClusterID)
		util.KubectlUseContext(cmdClusterID)

		fmt.Printf("Switched to kubectl context 'giantswarm-%s'\n\n", cmdClusterID)

		// final success message
		color.Green("kubectl is set up. Check it using this command:\n\n")
		color.Yellow("    kubectl cluster-info\n\n")
		color.Green("Whenever you want to switch to using this context:\n\n")
		color.Yellow("    kubectl config set-context giantswarm-%s\n\n", cmdClusterID)

	} else {
		fmt.Printf("Unhandled response code: %v", keypairResponse.StatusCode)
		fmt.Printf("Status text: %v", keypairResponse.StatusText)
	}
}
