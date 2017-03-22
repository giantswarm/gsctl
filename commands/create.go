package commands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/phayes/permbits"
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

	pkcs12BundlePassword = "giantswarm"
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

// checkOpenSSL returns true if openssl executable exists and is callable
func checkOpenSSL() bool {
	cmd := exec.Command("openssl")
	if err := cmd.Run(); err != nil {
		var waitStatus syscall.WaitStatus
		exitStatus := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			exitStatus = waitStatus.ExitStatus()
		}
		if exitStatus == 0 {
			return true
		}
		return false
	}
	return true
}

// executable checks whether there is an executable file at the given path
func executable(path string) (bool, error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		// does not exist
		return false, nil
	} else if err != nil {
		// some other error
		return false, err
	}
	if stat.Mode().IsDir() {
		// file is a directory
		return false, nil
	}
	// evaluate permissions
	permissions, permErr := permbits.Stat(path)
	if permErr != nil {
		return false, permErr
	}
	if permissions.UserExecute() || permissions.GroupExecute() {
		return true, nil
	}
	return false, err
}

// which checks if an executable is available in the PATH.
// If yes, the path is returned. Else, an empty string is return.
func which(name string) (string, error) {
	// iterate path
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, dirPath := range paths {
		execPath := dirPath + string(os.PathSeparator) + name
		isExecutable, err := executable(execPath)
		if err != nil {
			return "", err
		}
		if isExecutable {
			return execPath, nil
		}
	}
	return "", nil
}

func checkExecutableExists(name string) (bool, error) {
	myPath, err := which(name)
	if err != nil {
		return false, err
	}
	if myPath != "" {
		return true, nil
	}
	return false, nil
}

// firefoxProfiles returns a slice of firefox profile folders of the current user
func firefoxProfiles() []string {
	profiles := []string{}
	ffpath := path.Join(config.SystemUser.HomeDir, ".mozilla", "firefox")
	files, err := ioutil.ReadDir(ffpath)
	if err == nil {
		for _, f := range files {
			if strings.Contains(f.Name(), ".default") {
				// ff profile folder names end in ".default"
				profiles = append(profiles, path.Join(ffpath, f.Name()))
			}
		}
	}
	return profiles
}

// checkFirefoxProfileExists returns true if the user has a Firefox profile folder
func checkFirefoxProfileExists() bool {
	p := firefoxProfiles()
	if len(p) > 0 {
		return true
	}
	return false
}

// creates a PKCS#12 bundle from private RSA key and X.509 certificate using openssl
func createPkcs12Bundle(clusterID, keyPairID string) (string, error) {
	truncatedKeyPairID := util.CleanKeypairID(keyPairID)[:10]
	// TODO: check each file's existence separately and give meaningful output in case of error
	privateKeyPath := path.Join(config.CertsDirPath, clusterID+"-"+truncatedKeyPairID+"-client.key")
	clientCertificatePath := path.Join(config.CertsDirPath, clusterID+"-"+truncatedKeyPairID+"-client.crt")
	outputFilePath := path.Join(config.CertsDirPath, clusterID+"-"+truncatedKeyPairID+".p12")
	bundleName := "\"Giant Swarm " + clusterID + "-" + truncatedKeyPairID + "\""
	cmd := exec.Command("openssl",
		"pkcs12", "-export", "-clcerts",
		"-inkey", privateKeyPath,
		"-in", clientCertificatePath,
		"-out", outputFilePath,
		"-passout", "pass:"+pkcs12BundlePassword,
		"-name", bundleName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return outputFilePath, nil
	}
	fmt.Println(cmd.Stdout)
	fmt.Println(cmd.Stderr)
	return outputFilePath, err
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

// importCaAndPkcs12Bundle attempts to import the CA file and
// PKCS#12 bundle (.p12 file) to the available
// key chains or certificate databases.
// @returns []string Information where bundle has been imported to
func importCaAndPkcs12Bundle(caFilePath, p12path, clusterID, password string) []string {
	successfulImports := []string{}
	certName := fmt.Sprintf("Giant Swarm cluster '%s' CA", clusterID)
	u, _ := user.Current()

	// Mac only: import to OS keychain
	if runtime.GOOS == "darwin" {

		// CA
		keychainPath := path.Join(u.HomeDir, "Library", "Keychains", "login.keychain")
		cmd1 := exec.Command("security", "add-trusted-cert",
			"-r", "trustRoot",
			"-k", keychainPath,
			caFilePath)
		var out1 bytes.Buffer
		var stderr1 bytes.Buffer
		cmd1.Stdout = &out1
		cmd1.Stderr = &stderr1
		err1 := cmd1.Run()
		if err1 == nil {
			successfulImports = append(successfulImports, "CA to Mac OS system keychain")
		} else {
			fmt.Println("Importing CA to Mac OS keychain failed.")
			fmt.Printf("Details: %s\n", stderr1.String())
		}

		// PKCS#12
		if p12path != "" {
			cmd2 := exec.Command("security", "import", p12path, "-P", password)
			var out2 bytes.Buffer
			var stderr2 bytes.Buffer
			cmd2.Stdout = &out2
			cmd2.Stderr = &stderr2
			err2 := cmd2.Run()
			if err2 == nil {
				successfulImports = append(successfulImports, "PKCS#12 bundle to Mac OS system keychain")
			} else {
				fmt.Println("Importing PKCS#12 bundle to Mac OS keychain failed.")
				fmt.Printf("Details: %s\n", stderr2.String())
			}
		}
	}

	// Shared NSS database (e. g. Chrome, Chromium on Linux)
	sharedNssDbPath := path.Join(u.HomeDir, ".pki", "nssdb")
	if _, statErr := os.Stat(sharedNssDbPath); statErr == nil {
		if cmdVerbose {
			fmt.Printf("Shared NSS database exists in '%s'\n", sharedNssDbPath)
		}

		// CA
		if certutilExists, _ := checkExecutableExists("certutil"); certutilExists {
			fmt.Println("Please close all windows of browsers using your shared NSS database, like")
			fmt.Println("Chromium and Google Chrome, before you proceed.")
			fmt.Print("Press 'Enter' to continue")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			cmd3 := exec.Command("certutil", "-A",
				"-n", "\""+certName+"\"",
				"-t", "\"TC,,\"",
				"-d", "'sql:"+sharedNssDbPath+"'",
				"-i", caFilePath)
			var out3 bytes.Buffer
			var stderr3 bytes.Buffer
			cmd3.Stdout = &out3
			cmd3.Stderr = &stderr3
			err3 := cmd3.Run()
			if err3 == nil {
				successfulImports = append(successfulImports, "CA to shared NSS database")
			} else {
				fmt.Println("Importing CA to shared NSS database failed.")
				fmt.Printf("Details: %s\n", stderr3.String())
			}
		}

		// PKCS#12
		if p12path != "" {
			if pk12utilExists, _ := checkExecutableExists("pk12util"); pk12utilExists {
				cmd4 := exec.Command("pk12util",
					"-i", p12path,
					"-d", "'sql:"+sharedNssDbPath+"'",
					"-W", password)
				var out4 bytes.Buffer
				var stderr4 bytes.Buffer
				cmd4.Stdout = &out4
				cmd4.Stderr = &stderr4
				err4 := cmd4.Run()
				if err4 == nil {
					successfulImports = append(successfulImports, "PKCS#12 bundle to shared NSS database")
				} else {
					fmt.Println("Importing PKCS#12 bundle to shared NSS database failed.")
					fmt.Printf("Details: %s\n", stderr4.String())
				}
			} else {
				fmt.Println("Importing PKCS#12 bundle to shared NSS database failed.")
				fmt.Println("Executable 'pk12util' not found.")
				if runtime.GOOS == "linux" {
					fmt.Println("Debian-based distributions allow the installation of 'pk12util' via")
					fmt.Println("'sudo apt-get install libnss3-tools'.")
				}
			}
		}
	}

	// Firefox
	if checkFirefoxProfileExists() {

		if runtime.GOOS == "darwin" {
			// Mac OS
			fmt.Println("To import the CA and PKCS#12 bundle into Firefox,")
			fmt.Println("please refer to our guide at\n")
			fmt.Println("   https://docs.giantswarm.io/guides/importing-certificates/#mac-os-firefox\n")
		} else if runtime.GOOS == "windows" {
			// Windows
			fmt.Println("To import the CA and PKCS#12 bundle into Firefox,")
			fmt.Println("please open the Firefox certificate manager under")
			fmt.Println("'Tools' > 'Options' > 'Advanced' > 'Certificates'")
			fmt.Println("and use the function 'View Certificates' > 'Import'.")
		} else {
			// Linux or the likes (hopefully)
			fmt.Println("Please close all windows of Firefox before you proceed.")
			fmt.Print("Press 'Enter' to continue")
			bufio.NewReader(os.Stdin).ReadBytes('\n')

			// install in all profile folders
			for _, profileFolder := range firefoxProfiles() {
				// CA
				if certutilExists, _ := checkExecutableExists("certutil"); certutilExists {
					cmd5 := exec.Command("certutil", "-A",
						"-n", "\""+certName+"\"",
						"-t", "\"TC,,\"",
						"-d", "\""+profileFolder+"\"",
						"-i", caFilePath)
					var out5 bytes.Buffer
					var stderr5 bytes.Buffer
					cmd5.Stdout = &out5
					cmd5.Stderr = &stderr5
					err5 := cmd5.Run()
					if err5 == nil {
						successfulImports = append(successfulImports, fmt.Sprintf("CA to Firefox profile %s", profileFolder))
					} else {
						fmt.Printf("Importing CA to Firefox profile %s failed.\n", profileFolder)
						fmt.Printf("Details: %s\n", stderr5.String())
					}
				} else {
					fmt.Println("Importing CA to Firefox failed.")
					fmt.Println("Executable 'certutil' not found.")
					fmt.Println("Debian-based Linux distributions allow the installation of 'pk12util' via")
					fmt.Println("'sudo apt-get install libnss3-tools'.")
				}

				// PKCS#12
				if p12path != "" {
					if pk12utilExists, _ := checkExecutableExists("pk12util"); pk12utilExists {
						cmd6 := exec.Command("pk12util",
							"-i", p12path,
							"-d", "\""+profileFolder+"\"",
							"-W", password)
						var out6 bytes.Buffer
						var stderr6 bytes.Buffer
						cmd6.Stdout = &out6
						cmd6.Stderr = &stderr6
						err6 := cmd6.Run()
						if err6 == nil {
							successfulImports = append(successfulImports, fmt.Sprintf("PKCS#12 bundle to Firefox profile %s", profileFolder))
						} else {
							fmt.Printf("Importing PKCS#12 bundle to Firefox profile %s failed.\n", profileFolder)
							fmt.Printf("Details: %s\n", stderr6.String())
						}
					} else {
						fmt.Println("Importing PKCS#12 bundle to Firefox failed.")
						fmt.Println("Executable 'pk12util' not found.")
						fmt.Println("Debian-based Linux distributions allow the installation of 'pk12util' via")
						fmt.Println("'sudo apt-get install libnss3-tools'.")
					}
				}

			}

		}
	}

	return successfulImports
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
	var p12Path string

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

	keypairResponse, apiResponse, err := client.AddKeyPair(authHeader, cmdClusterID, addKeyPairBody, requestIDHeader, createKubeconfigActivityName, cmdLine)

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

		// Create PKCS#12 bundle
		if checkOpenSSL() {
			thisP12Path, thisErr := createPkcs12Bundle(cmdClusterID, keypairResponse.Data.Id)
			if thisErr != nil {
				fmt.Println(color.YellowString("Warning: Could not create PKCS#12 bundle."))
				fmt.Println(color.YellowString("Details: %v", thisErr))
			} else {
				p12Path = thisP12Path
				fmt.Println(fmt.Sprintf("Created PKCS#12 bundle with default password '%s' in:", pkcs12BundlePassword))
				fmt.Println(p12Path)
			}
		} else {
			fmt.Println(color.YellowString("Info: OpenSSL executable could not be found. PKCS#12 bundle has not been created."))
			fmt.Println(color.YellowString("Refer to\n"))
			fmt.Println(color.YellowString("    https://docs.giantswarm.io/guides/importing-certificates/\n"))
			fmt.Println(color.YellowString("for details on how to establish trust to your cluster"))
			fmt.Println(color.YellowString("and import your key pair into a browser.\n"))
		}

		imported := importCaAndPkcs12Bundle(caCertPath, p12Path, cmdClusterID, pkcs12BundlePassword)
		if len(imported) > 0 {
			fmt.Println("Successful certificate imports:")
			for n, line := range imported {
				fmt.Printf(" %d. %s\n", (n + 1), line)
			}
			fmt.Println("")
		}

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
