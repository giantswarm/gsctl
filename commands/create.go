package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
	// CreateCommand is the command to create things
	CreateCommand = &cobra.Command{
		Use:   "create",
		Short: "Create clusters, key-pairs, ...",
		Long:  `Lets you create things like clusters, key-pairs or kubectl configuration files`,
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

const (
	// url to intallation instructions
	kubectlInstallURL string = "http://kubernetes.io/docs/user-guide/prereqs/"

	// windows download page
	kubectlWindowsInstallURL string = "https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md"

	addKeyPairActivityName       string = "add-keypair"
	createKubeconfigActivityName string = "create-kubeconfig"
)

func init() {
	CreateKeypairCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key-pair for")
	CreateKeypairCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")

	CreateKubeconfigCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")
	CreateKubeconfigCommand.Flags().IntVarP(&cmdTTLDays, "ttl", "", 30, "Duration until expiry of the created key-pair in days")

	// subcommands
	CreateCommand.AddCommand(CreateKeypairCommand, CreateKubeconfigCommand)

	RootCommand.AddCommand(CreateCommand)
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

// Pre-check before creating a new kubeconfig
func checkCreateKubeconfig(cmd *cobra.Command, args []string) error {
	kubectlOkay := util.CheckKubectl()
	if !kubectlOkay {
		// kubectl not installed
		errorMessage := color.RedString("kubectl does not appear to be installed") + "\n"
		if runtime.GOOS == "darwin" {
			errorMessage += "Please install via 'brew install kubernetes-cli' or visit\n"
			errorMessage += fmt.Sprintf("%s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "linux" {
			errorMessage += fmt.Sprintf("Please visit %s for information on how to install kubectl", kubectlInstallURL)
		} else if runtime.GOOS == "windows" {
			errorMessage += fmt.Sprintf("Please visit %s to download a recent kubectl binary.", kubectlWindowsInstallURL)
		}
		return errors.New(errorMessage)
	}

	if config.Config.Token == "" {
		return errors.New("You are not logged in. Use '" + config.ProgramName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := config.GetDefaultCluster(requestIDHeader, createKubeconfigActivityName, cmdLine, cmdAPIEndpoint)
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
	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	authHeader := "giantswarm " + config.Config.Token

	// get cluster details
	clusterDetailsResponse, apiResponse, err := client.GetCluster(authHeader, cmdClusterID, requestIDHeader, createKubeconfigActivityName, cmdLine)
	if err != nil {
		fmt.Println(color.RedString("Could not fetch details for cluster ID '" + cmdClusterID + "'"))
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}

	// parameters given by the user
	ttlHours := int32(cmdTTLDays * 24)
	if cmdDescription == "" {
		cmdDescription = "Added by user " + config.Config.Email + " using 'gsctl create kubeconfig'"
	}

	addKeyPairBody := gsclientgen.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}

	fmt.Println("Creating new key-pair…")

	keypairResponse, _, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody, requestIDHeader, createKubeconfigActivityName, cmdLine)

	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}

	if keypairResponse.StatusCode == apischema.STATUS_CODE_DATA {
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

		// edit kubectl config
		if err := util.KubectlSetCluster(cmdClusterID, clusterDetailsResponse.ApiEndpoint, caCertPath); err != nil {
			fmt.Println(color.RedString("Could not set cluster using 'kubectl config set-cluster ...'"))
			fmt.Println("Error:")
			fmt.Println(err)
			os.Exit(1)
		}

		if err := util.KubectlSetCredentials(cmdClusterID, clientKeyPath, clientCertPath); err != nil {
			fmt.Println(color.RedString("Could not set credentials using 'kubectl config set-credentials ...'"))
			fmt.Println("Error:")
			fmt.Println(err)
			os.Exit(1)
		}

		if err := util.KubectlSetContext(cmdClusterID); err != nil {
			fmt.Println(color.RedString("Could not set context using 'kubectl config set-context ...'"))
			fmt.Println("Error:")
			fmt.Println(err)
			os.Exit(1)
		}

		if err := util.KubectlUseContext(cmdClusterID); err != nil {
			fmt.Println(color.RedString("Could not apply context using 'kubectl config use-context giantswarm-%s'", cmdClusterID))
			fmt.Println("Error:")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Switched to kubectl context 'giantswarm-%s'\n\n", cmdClusterID)

		// final success message
		fmt.Println(color.GreenString("kubectl is set up. Check it using this command:\n"))
		fmt.Println(color.YellowString("    kubectl cluster-info\n"))
		fmt.Println(color.GreenString("Whenever you want to switch to using this context:\n"))
		fmt.Println(color.YellowString("    kubectl config set-context giantswarm-%s\n", cmdClusterID))

	} else {
		fmt.Println(color.RedString("Unhandled response code: %v", keypairResponse.StatusCode))
		dumpAPIResponse(*apiResponse)
	}
}
