package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	goclient "github.com/giantswarm/go-client-gen"
	"github.com/howeyc/gopass"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/util"
)

var (
	// to be set on build by the go linker
	version   string
	buildDate string

	// SystemUser is the current user of the system
	SystemUser *user.User

	// The password used for authentication
	password string

	err error

	// path to our own configuration directory
	configDirPath string

	// paths to known kubeconfig files
	kubeConfigPaths []string

	myConfig *Config

	// flags to be used based on commands
	cmdClusterID   string
	cmdDescription string
	cmdVerbose     bool
	cmdTTLDays     int
	cmdToken       string
)

const (
	// name of the configuration file
	configFileName string = "config.yaml"
	// name of this program
	programName string = "gsctl"
)

// Config is the data structure for our configuration
type Config struct {
	Token        string
	Email        string
	Updated      string
	Organization string
	Cluster      string
}

// printInfo prints some information on the current user
func printInfo(cmd *cobra.Command, args []string) {
	output := []string{}

	if myConfig.Organization == "" {
		output = append(output, color.YellowString("Selected organization:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Selected organization:")+"|"+color.CyanString(myConfig.Organization))
	}

	if myConfig.Cluster == "" {
		output = append(output, color.YellowString("Selected cluster:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Selected cluster:")+"|"+color.CyanString(myConfig.Cluster))
	}

	if myConfig.Email == "" {
		output = append(output, color.YellowString("Email:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Email:")+"|"+color.CyanString(myConfig.Email))
	}

	if cmdToken != "" {
		output = append(output, color.YellowString("Auth token:")+"|"+color.CyanString(cmdToken))
	} else if myConfig.Token != "" {
		output = append(output, color.YellowString("Auth token:")+"|"+color.CyanString(myConfig.Token))
	} else {
		output = append(output, color.YellowString("Auth token:")+"|n/a")
	}

	output = append(output, color.YellowString("g8m version:")+"|"+version)
	output = append(output, color.YellowString("g8m build:")+"|"+buildDate)

	output = append(output, color.YellowString("Config path:")+"|"+color.CyanString(configDirPath))

	// kubectl configuration paths
	if len(kubeConfigPaths) == 0 {
		output = append(output, color.YellowString("kubectl config path:")+"|n/a")
	} else {
		paths := []string{}
		for _, myPath := range kubeConfigPaths {
			paths = append(paths, color.CyanString(myPath))
		}
		output = append(output, color.YellowString("kubectl config path:")+"|"+strings.Join(paths, ", "))
	}

	fmt.Println(columnize.SimpleFormat(output))
}

// ping checks the API connections
func ping(cmd *cobra.Command, args []string) {
	uri := "https://api.giantswarm.io/v1/ping"
	start := time.Now()
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(color.RedString("API cannot be reached"))
		log.Fatal(err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		elapsed := time.Since(start)
		if resp.StatusCode == 200 {
			fmt.Println(color.GreenString("API connection is fine"))
			fmt.Printf("Ping took %v\n", elapsed)
		} else {
			fmt.Println(color.RedString("API returned unexpected response status %v", resp.StatusCode))
			os.Exit(2)
		}
	}
}

// checks if all arguments for the login command are given
func checkLogin(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("The email argument is required")
	}

	// using auth token flag?
	if cmdToken != "" {
		return errors.New("The 'login' command cannot be used with the '--auth-token' flag")
	}

	// already logged in?
	if myConfig.Token != "" {
		return errors.New("You are already logged in")
	}

	return nil
}

// login creates a new session token
func login(cmd *cobra.Command, args []string) {
	var email = args[0]

	// interactive password prompt
	if password == "" {
		fmt.Printf("Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			log.Fatal(err)
		}
		password = string(pass)
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

	client := goclient.NewDefaultApi()
	requestBody := goclient.LoginBody{Password: string(encodedPassword)}
	loginResponse, apiResponse, err := client.UserLogin(email, requestBody)
	if err != nil {
		log.Fatal(err)
	}
	if loginResponse.StatusCode == 10000 {
		// successful login
		fmt.Println(color.GreenString("Successfully logged in"))
		myConfig.Token = loginResponse.Data.Id
		myConfig.Email = email
	} else if loginResponse.StatusCode == 10010 {
		// bad credentials
		fmt.Println(color.RedString("Incorrect password submitted. Please try again."))
	} else {
		fmt.Printf("Unhandled response code: %v", loginResponse.StatusCode)
		fmt.Printf("Status text: %v", loginResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
	}
}

func checkLogout(cmd *cobra.Command, args []string) error {
	if myConfig.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in")
	}
	return nil
}

func logout(cmd *cobra.Command, args []string) {
	client := goclient.NewDefaultApi()

	// if token is set via flags, we unauthenticate using this token
	authHeader := "giantswarm " + myConfig.Token
	if cmdToken != "" {
		authHeader = "giantswarm " + cmdToken
	}

	logoutResponse, apiResponse, err := client.UserLogout(authHeader)
	if err != nil {
		fmt.Println("Info: The client doesn't handle the API's 401 response yet.")
		fmt.Println("Seeing this error likely means: The passed token was no longer valid.")
		fmt.Println("Error details:")
		log.Fatal(err)
	}
	if logoutResponse.StatusCode == 10007 {
		// remove token from settings
		// unless we unathenticated the token from flags
		if cmdToken == "" {
			myConfig.Token = ""
			myConfig.Email = ""
		}
		fmt.Println(color.GreenString("Successfully logged out"))
	} else {
		fmt.Printf("Unhandled response code: %v", logoutResponse.StatusCode)
		fmt.Printf("Status text: %v", logoutResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
	}
}

func checkListOrgs(cmd *cobra.Command, args []string) error {
	if myConfig.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in.\nUse '" + programName + " login' to login or '--auth-token' to pass a valid auth token.")
	}
	return nil
}

// list organizations the user is member of
func listOrgs(cmd *cobra.Command, args []string) {
	client := goclient.NewDefaultApi()

	// if token is set via flags, we unauthenticate using this token
	authHeader := "giantswarm " + myConfig.Token
	if cmdToken != "" {
		authHeader = "giantswarm " + cmdToken
	}

	orgsResponse, apiResponse, err := client.GetUserOrganizations(authHeader)
	if err != nil {
		fmt.Println("Error details:")
		log.Fatal(err)
	}

	if orgsResponse.StatusCode == 10000 {
		var organizations = orgsResponse.Data
		if len(organizations) == 0 {
			fmt.Println(color.YellowString("No organizations available"))
		} else {
			sort.Strings(organizations)
			fmt.Println(color.YellowString("Organization"))
			for _, orgName := range organizations {
				fmt.Println(color.CyanString(orgName))
			}
		}
	} else {
		fmt.Printf("Unhandled response code: %v", orgsResponse.StatusCode)
		fmt.Printf("Status text: %v", orgsResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
	}
}

func checkListClusters(cmd *cobra.Command, args []string) error {
	if myConfig.Token == "" {
		return errors.New("You are not logged in. Use '" + programName + " login' to log in.")
	}
	return nil
}

// list all clusters the user has access to
func listClusters(cmd *cobra.Command, args []string) {
	client := goclient.NewDefaultApi()
	authHeader := "giantswarm " + myConfig.Token
	orgsResponse, apiResponse, err := client.GetUserOrganizations(authHeader)
	if err != nil {
		log.Fatal(err)
	}
	if orgsResponse.StatusCode == 10000 {
		var organizations = orgsResponse.Data
		if len(organizations) == 0 {
			fmt.Println(color.YellowString("No organizations available"))
		} else {
			sort.Strings(organizations)
			output := []string{color.YellowString("Id") + "|" + color.YellowString("Name") + "|" + color.YellowString("Created") + "|" + color.YellowString("Organization")}
			for _, orgName := range organizations {
				clustersResponse, _, err := client.GetOrganizationClusters(authHeader, orgName)
				if err != nil {
					log.Fatal(err)
				}
				for _, cluster := range clustersResponse.Data.Clusters {
					created := util.ShortDate(util.ParseDate(cluster.CreateDate))
					output = append(output,
						color.CyanString(cluster.Id)+"|"+
							color.CyanString(cluster.Name)+"|"+
							color.CyanString(created)+"|"+
							color.CyanString(orgName))
				}
			}
			fmt.Println(columnize.SimpleFormat(output))
		}
	} else {
		fmt.Printf("Unhandled response code: %v", orgsResponse.StatusCode)
		fmt.Printf("Status text: %v", orgsResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
	}
}

func checkListKeypairs(cmd *cobra.Command, args []string) error {
	if myConfig.Token == "" {
		return errors.New("You are not logged in. Use '" + programName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := getDefaultCluster()
		if clusterID != "" {
			cmdClusterID = clusterID
		} else {
			return errors.New("No cluster given. Please use the -c/--cluster flag to set a cluster ID.")
		}
	}
	return nil
}

func listKeypairs(cmd *cobra.Command, args []string) {
	client := goclient.NewDefaultApi()
	authHeader := "giantswarm " + myConfig.Token
	keypairsResponse, _, err := client.GetKeyPairs(authHeader, cmdClusterID)
	if err != nil {
		log.Fatal(err)
	}
	if keypairsResponse.StatusCode == 10000 {
		// sort result
		slice.Sort(keypairsResponse.Data.KeyPairs[:], func(i, j int) bool {
			return keypairsResponse.Data.KeyPairs[i].CreateDate < keypairsResponse.Data.KeyPairs[j].CreateDate
		})

		// create output
		output := []string{color.YellowString("Created") + "|" + color.YellowString("Expires") + "|" + color.YellowString("Id") + "|" + color.YellowString("Description")}
		for _, keypair := range keypairsResponse.Data.KeyPairs {
			created := util.ShortDate(util.ParseDate(keypair.CreateDate))
			expires := util.ParseDate(keypair.CreateDate).Add(time.Duration(keypair.TtlHours) * time.Hour)

			// skip if expired
			output = append(output, color.CyanString(created)+"|"+
				color.CyanString(util.ShortDate(expires))+"|"+
				color.CyanString(util.Truncate(util.CleanKeypairID(keypair.Id), 10))+"|"+
				color.CyanString(keypair.Description))
		}
		fmt.Println(columnize.SimpleFormat(output))

	} else {
		fmt.Printf("Unhandled response code: %v", keypairsResponse.StatusCode)
		fmt.Printf("Status text: %v", keypairsResponse.StatusText)
	}
}

func checkAddKeypair(cmd *cobra.Command, args []string) error {
	if myConfig.Token == "" {
		return errors.New("You are not logged in. Use '" + programName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := getDefaultCluster()
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
	client := goclient.NewDefaultApi()
	authHeader := "giantswarm " + myConfig.Token
	// TODO: Make ttl configurable
	ttlHours := int32(3)
	addKeyPairBody := goclient.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}
	keypairResponse, _, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody)

	if err != nil {
		log.Fatal(err)
	}

	if keypairResponse.StatusCode == 10000 {
		cleanID := util.CleanKeypairID(keypairResponse.Data.Id)
		msg := fmt.Sprintf("New key-pair created with ID %s", cleanID)
		fmt.Println(color.GreenString(msg))

		// store credentials to file
		caCertPath := util.StoreCaCertificate(configDirPath, cmdClusterID, keypairResponse.Data.CertificateAuthorityData)
		fmt.Println("CA certificate stored in:", caCertPath)

		clientCertPath := util.StoreClientCertificate(configDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientCertificateData)
		fmt.Println("Client certificate stored in:", clientCertPath)

		clientKeyPath := util.StoreClientKey(configDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientKeyData)
		fmt.Println("Client private key stored in:", clientKeyPath)

	} else {
		fmt.Printf("Unhandled response code: %v", keypairResponse.StatusCode)
		fmt.Printf("Status text: %v", keypairResponse.StatusText)
	}
}

// Pre-check before creating a new kubeconfig
func checkCreateKubeconfig(cmd *cobra.Command, args []string) error {
	util.CheckKubectl()
	if myConfig.Token == "" {
		return errors.New("You are not logged in. Use '" + programName + " login' to log in.")
	}
	if cmdClusterID == "" {
		// use default cluster if possible
		clusterID, _ := getDefaultCluster()
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
	client := goclient.NewDefaultApi()
	authHeader := "giantswarm " + myConfig.Token

	// parameters given by the user
	ttlHours := int32(cmdTTLDays * 24)
	if cmdDescription == "" {
		cmdDescription = "Added by user " + myConfig.Email + " using 'g8m create kubeconfig'"
	}

	addKeyPairBody := goclient.AddKeyPairBody{Description: cmdDescription, TtlHours: ttlHours}

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
		caCertPath := util.StoreCaCertificate(configDirPath, cmdClusterID, keypairResponse.Data.CertificateAuthorityData)

		clientCertPath := util.StoreClientCertificate(configDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientCertificateData)

		clientKeyPath := util.StoreClientKey(configDirPath, cmdClusterID, keypairResponse.Data.Id, keypairResponse.Data.ClientKeyData)

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

// writes configuration to YAML file
func writeConfig() {

	// ensure directory
	os.MkdirAll(configDirPath, 0700)

	var configFilePath = configDirPath + string(os.PathSeparator) + configFileName

	// last modified date
	myConfig.Updated = time.Now().Format(time.RFC3339)

	yamlBytes, yamlErr := yaml.Marshal(&myConfig)
	if yamlErr != nil {
		log.Fatal(yamlErr)
	}
	err := ioutil.WriteFile(configFilePath, yamlBytes, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

// read configuration from YAML file
func readConfig() {
	var configFilePath = configDirPath + string(os.PathSeparator) + configFileName

	data, readErr := ioutil.ReadFile(configFilePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return
		}
		log.Fatal(readErr)
	}

	yamlErr := yaml.Unmarshal(data, &myConfig)
	if yamlErr != nil {
		log.Fatal(yamlErr)
	}
}

// getKubeconfigPaths returns a alice of paths to known kubeconfig files.
func getKubeconfigPaths() []string {

	// check if KUBECONFIG environment variable is set
	kubeconfigEnv := os.Getenv("KUBECONFIG")

	if kubeconfigEnv != "" {
		// KUBECONFIG is set.
		// Now check all the paths included for file existence
		pathSep := ":"
		if runtime.GOOS == "windows" {
			pathSep = ";"
		}
		paths := strings.Split(kubeconfigEnv, pathSep)
		out := []string{}
		for _, myPath := range paths {
			if _, err := os.Stat(myPath); err == nil {
				out = append(out, myPath)
			}
		}
		return out
	}

	// KUBECONFIG is not set.
	// Look for the default location ~/.kube/config
	kubeconfigPath := path.Join(SystemUser.HomeDir, ".kube", "config")
	if _, err := os.Stat(kubeconfigPath); err == nil {
		// file exists in default location
		return []string{kubeconfigPath}
	}

	// No kubeconfig file. Return empty slice.
	return []string{}
}

// determine which is the default cluster. This can be either
// the only cluster accessible, or a cluster selected explicitly.
func getDefaultCluster() (clusterID string, err error) {
	// Check selected cluster
	if myConfig.Cluster != "" {
		return myConfig.Cluster, nil
	}
	// Go through available orgs and clusters to find all clusters
	if myConfig.Token == "" {
		return "", errors.New("User not logged in.")
	}
	client := goclient.NewDefaultApi()
	authHeader := "giantswarm " + myConfig.Token
	orgsResponse, _, err := client.GetUserOrganizations(authHeader)
	if err != nil {
		return "", err
	}
	if orgsResponse.StatusCode == 10000 {
		if len(orgsResponse.Data) > 0 {
			clusterIDs := []string{}
			for _, orgName := range orgsResponse.Data {
				clustersResponse, _, err := client.GetOrganizationClusters(authHeader, orgName)
				if err != nil {
					return "", err
				}
				for _, cluster := range clustersResponse.Data.Clusters {
					clusterIDs = append(clusterIDs, cluster.Id)
				}
			}
			if len(clusterIDs) == 1 {
				return clusterIDs[0], nil
			}
			return "", nil
		}
	}
	return "", errors.New(orgsResponse.StatusText)
}

func main() {

	// initialize some global variables
	myConfig = new(Config)
	SystemUser, err = user.Current()
	if err != nil {
		log.Fatal(err)
	}
	configDirPath = SystemUser.HomeDir + string(os.PathSeparator) + "." + programName

	// Overwrite myConfig from config file (if exists)
	readConfig()

	columnizeConfig := columnize.DefaultConfig()
	columnizeConfig.Glue = "   "

	// disable color on windows, as it is super slow
	if runtime.GOOS == "windows" {
		color.NoColor = true
	}

	kubeConfigPaths = getKubeconfigPaths()

	var cmdPing = &cobra.Command{
		Use:   "ping",
		Short: "Check API connection",
		Long:  `Tests the connection to the API`,
		Run:   ping,
	}

	var cmdInfo = &cobra.Command{
		Use:   "info",
		Short: "Print some information",
		Long:  `Prints information that might help you get out of trouble`,
		Run:   printInfo,
	}

	var cmdLogin = &cobra.Command{
		Use:     "login <email>",
		Short:   "Sign in as a user",
		Long:    `Sign in with email address and password. Password has to be entered interactively or given as -p flag.`,
		PreRunE: checkLogin,
		Run:     login,
	}
	cmdLogin.Flags().StringVarP(&password, "password", "p", "", "Password. If not given, will be prompted interactively.")

	var cmdLogout = &cobra.Command{
		Use:     "logout",
		Short:   "Sign out the current user",
		Long:    `This will terminate the current user's session and invalidate the authentication token.`,
		PreRunE: checkLogout,
		Run:     logout,
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List things, like organizations, clusters, key-pairs",
		Long:  `Prints a list of the things you have access to`,
	}

	var cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "Create things, like key-pairs",
		Long:  `Lets you create things like key-pairs`,
	}

	var cmdListOrgs = &cobra.Command{
		Use:     "organizations",
		Short:   "List organizations",
		Long:    `Prints a list of the organizations you are a member of`,
		PreRunE: checkListOrgs,
		Run:     listOrgs,
	}

	var cmdListClusters = &cobra.Command{
		Use:     "clusters",
		Short:   "List clusters",
		Long:    `Prints a list of all clusters you have access to`,
		PreRunE: checkListClusters,
		Run:     listClusters,
	}

	var cmdListKeypairs = &cobra.Command{
		Use:     "keypairs",
		Short:   "List key-pairs for a cluster",
		Long:    `Prints a list of key-pairs for a cluster`,
		PreRunE: checkListKeypairs,
		Run:     listKeypairs,
	}
	cmdListKeypairs.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to list key-pairs for")

	var cmdCreateKeypair = &cobra.Command{
		Use:     "keypair",
		Short:   "Create key-pair",
		Long:    `Creates a new key-pair for a cluster`,
		PreRunE: checkAddKeypair,
		Run:     addKeypair,
	}
	cmdCreateKeypair.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster to create a key-pair for")
	cmdCreateKeypair.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")

	var cmdCreateKubeconfig = &cobra.Command{
		Use:     "kubeconfig",
		Short:   "Configure kubectl",
		Long:    `Modifies kubectl configuration to access your Giant Swarm Kubernetes cluster`,
		PreRunE: checkCreateKubeconfig,
		Run:     createKubeconfig,
	}
	cmdCreateKubeconfig.Flags().StringVarP(&cmdClusterID, "cluster", "c", "", "ID of the cluster")
	cmdCreateKubeconfig.Flags().StringVarP(&cmdDescription, "description", "d", "", "Description for the key-pair")
	cmdCreateKubeconfig.Flags().IntVarP(&cmdTTLDays, "ttl", "", 30, "Duration until expiry of the created key-pair in days")

	var rootCmd = &cobra.Command{Use: programName}
	rootCmd.PersistentFlags().StringVarP(&cmdToken, "auth-token", "", "", "Authorization token to use for one command execution")
	rootCmd.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")

	// subcommands of "list"
	cmdList.AddCommand(cmdListOrgs, cmdListClusters, cmdListKeypairs)

	// subcommands of "create"
	cmdCreate.AddCommand(cmdCreateKeypair, cmdCreateKubeconfig)

	// top level commands
	rootCmd.AddCommand(cmdList, cmdCreate, cmdInfo, cmdPing, cmdLogin, cmdLogout)
	rootCmd.Execute()

	writeConfig()

}
