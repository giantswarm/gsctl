package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/gsctl/util"
)

var (
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
	createKubeconfigActivityName = "create-kubeconfig"
)

func init() {
	CreateKubeconfigCommand.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key pair")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdCNPrefix, "cn-prefix", "", "", "The common name prefix for the issued certificates 'CN' field.")
	CreateKubeconfigCommand.Flags().StringVarP(&cmdCertificateOrganizations, "certificate-organizations", "", "", "A comma separated list of organizations for the issued certificates 'O' fields.")
	CreateKubeconfigCommand.Flags().IntVarP(&cmdTTLDays, "ttl", "", 30, "Duration until expiry of the created key pair in days")

	CreateCommand.AddCommand(CreateKubeconfigCommand)
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
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)

	clientConfig := client.Configuration{
		Endpoint:  endpoint,
		Timeout:   10 * time.Second,
		UserAgent: config.UserAgent(),
	}
	apiClient, clientErr := client.NewClient(clientConfig)
	if clientErr != nil {
		fmt.Println(color.RedString("Error: %s", clientErr.Error()))
		os.Exit(1)
	}
	authHeader := "giantswarm " + token

	// get cluster details
	clusterDetailsResponse, apiResponse, err := apiClient.GetCluster(authHeader, cmdClusterID, requestIDHeader, createKubeconfigActivityName, cmdLine)
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

	addKeyPairBody := gsclientgen.V4AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours, CnPrefix: cmdCNPrefix, CertificateOrganizations: cmdCertificateOrganizations}

	fmt.Println("Creating new key pair…")

	clientConfig.Timeout = 60 * time.Second
	apiClient, clientErr = client.NewClient(clientConfig)
	if clientErr != nil {
		fmt.Println(color.RedString("Could not create API client.'"))
		fmt.Println("Error:")
		fmt.Println(clientErr)
		os.Exit(1)
	}

	keypairResponse, apiResponse, err := apiClient.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody, requestIDHeader, createKubeconfigActivityName, cmdLine)

	if err != nil {
		fmt.Println(color.RedString("Error: %s", err))
		dumpAPIResponse(*apiResponse)
		os.Exit(1)
	}

	if apiResponse.StatusCode == 200 || apiResponse.StatusCode == 201 {
		msg := fmt.Sprintf("New key pair created with ID %s and expiry of %v hours",
			util.Truncate(util.CleanKeypairID(keypairResponse.Id), 10),
			keypairResponse.TtlHours)
		fmt.Println(msg)

		// store credentials to file
		caCertPath := util.StoreCaCertificate(config.CertsDirPath, cmdClusterID, keypairResponse.CertificateAuthorityData)

		clientCertPath := util.StoreClientCertificate(config.CertsDirPath, cmdClusterID, keypairResponse.Id, keypairResponse.ClientCertificateData)

		clientKeyPath := util.StoreClientKey(config.CertsDirPath, cmdClusterID, keypairResponse.Id, keypairResponse.ClientKeyData)

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
		fmt.Println(color.YellowString("    kubectl config use-context giantswarm-%s\n", cmdClusterID))

	} else {
		fmt.Println(color.RedString("Unhandled response code: %v", apiResponse.StatusCode))
		dumpAPIResponse(*apiResponse)
	}
}
