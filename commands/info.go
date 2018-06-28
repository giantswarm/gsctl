package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/config"
)

const (
	infoActivityName = "info"
)

var (
	// InfoCommand is the "info" go command
	InfoCommand = &cobra.Command{
		Use:   "info",
		Short: "Print some information",
		Long:  `Prints information that might help you get out of trouble`,
		Run:   printInfo,
	}
)

// infoArguments represents the arguments we can make use of in this command
type infoArguments struct {
	scheme      string
	token       string
	verbose     bool
	apiEndpoint string
}

// defaultInfoArguments returns an infoArguments object populated by the user's
// command line arguments
func defaultInfoArguments() infoArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return infoArguments{
		scheme:      scheme,
		token:       token,
		verbose:     cmdVerbose,
		apiEndpoint: endpoint,
	}
}

// infoResult is the struct used to return all the info we might want to print
type infoResult struct {
	apiEndpoint      string
	apiEndpointAlias string
	email            string
	token            string
	version          string
	buildDate        string
	configFilePath   string
	kubeConfigPaths  []string
	infoResponse     *gsclientgen.V4InfoResponse
}

func init() {
	RootCommand.AddCommand(InfoCommand)
}

// printInfo prints some information on the current user and configuration
func printInfo(cmd *cobra.Command, args []string) {

	infoArgs := defaultInfoArguments()
	result, err := info(infoArgs)

	output := []string{}

	output = append(output, color.YellowString("%s version:", config.ProgramName)+"|"+color.CyanString(result.version))
	output = append(output, color.YellowString("%s build:", config.ProgramName)+"|"+color.CyanString(result.buildDate))
	output = append(output, color.YellowString("Config path:")+"|"+color.CyanString(result.configFilePath))

	// kubectl configuration paths
	output = append(output, color.YellowString("kubectl config path:")+"|"+color.CyanString(strings.Join(result.kubeConfigPaths, ", ")))

	if result.apiEndpoint == "" {
		output = append(output, color.YellowString("API endpoint:")+"|n/a")
	} else {
		output = append(output, color.YellowString("API endpoint:")+"|"+color.CyanString(result.apiEndpoint))
	}

	if result.apiEndpointAlias != "" {
		output = append(output, color.YellowString("API endpoint alias:")+"|"+color.CyanString(result.apiEndpointAlias))
	}

	if result.email == "" {
		output = append(output, color.YellowString("Email:")+"|n/a")
	} else {
		output = append(output, color.YellowString("Email:")+"|"+color.CyanString(config.Config.Email))
	}

	if result.token == "" {
		output = append(output, color.YellowString("Logged in:")+"|"+color.CyanString("no"))
	} else {
		output = append(output, color.YellowString("Logged in:")+"|"+color.CyanString("yes"))
	}

	if infoArgs.verbose {
		if result.token != "" {
			output = append(output, color.YellowString("Auth token:")+"|"+color.CyanString(result.token))
		} else {
			output = append(output, color.YellowString("Auth token:")+"|n/a")
		}
	}

	// Info depending on API communication
	if result.apiEndpoint != "" {
		// Provider
		if result.infoResponse == nil || result.infoResponse.General.Provider == "" {
			output = append(output, color.YellowString("Provider:")+"|n/a")
		} else {
			output = append(output, color.YellowString("Provider:")+"|"+color.CyanString(result.infoResponse.General.Provider))
		}

		if result.infoResponse != nil {
			if result.infoResponse.General.Provider == "aws" {
				output = append(output, color.YellowString("Worker instance type options:")+"|"+color.CyanString(strings.Join(result.infoResponse.Workers.InstanceType.Options, ", ")))
				output = append(output, color.YellowString("Default worker instance type:")+"|"+color.CyanString(result.infoResponse.Workers.InstanceType.Default_))
			} else if result.infoResponse.General.Provider == "azure" {
				output = append(output, color.YellowString("Worker VM size options:")+"|"+color.CyanString(strings.Join(result.infoResponse.Workers.VmSize.Options, ", ")))
				output = append(output, color.YellowString("Default worker VM size:")+"|"+color.CyanString(result.infoResponse.Workers.VmSize.Default_))
			}

			if result.infoResponse.Workers.CountPerCluster.Default_ != 0 {
				output = append(output, color.YellowString("Default workers per cluster:")+"|"+color.CyanString(fmt.Sprintf("%.0f", result.infoResponse.Workers.CountPerCluster.Default_)))
			}
			if result.infoResponse.Workers.CountPerCluster.Max != 0 {
				output = append(output, color.YellowString("Maximum workers per cluster:")+"|"+color.CyanString(fmt.Sprintf("%.0f", result.infoResponse.Workers.CountPerCluster.Max)))
			}
		}
	}

	fmt.Println(columnize.SimpleFormat(output))

	if err != nil {
		fmt.Println()
		fmt.Println(color.RedString("Some error occurred:"))
		fmt.Println(err.Error())
	}
}

// info gets all the information we'd like to show with the "info" command
// and returns it as a struct
func info(args infoArguments) (infoResult, error) {
	result := infoResult{}

	if args.apiEndpoint != "" {
		result.apiEndpoint = args.apiEndpoint
	}

	result.email = config.Config.Email
	result.token = config.Config.ChooseToken(result.apiEndpoint, args.token)
	result.version = config.Version
	result.buildDate = config.BuildDate

	if config.Config.Endpoints[result.apiEndpoint] != nil {
		result.apiEndpointAlias = config.Config.Endpoints[result.apiEndpoint].Alias
	}

	result.configFilePath = config.ConfigFilePath

	// kubectl configuration paths
	if len(config.KubeConfigPaths) > 0 {
		for _, myPath := range config.KubeConfigPaths {
			result.kubeConfigPaths = append(result.kubeConfigPaths, myPath)
		}
	}

	// get more info from API
	if args.apiEndpoint != "" {
		clientConfig := client.Configuration{
			Endpoint:  args.apiEndpoint,
			Timeout:   5 * time.Second,
			UserAgent: config.UserAgent(),
		}
		apiClient, clientErr := client.NewClient(clientConfig)
		if clientErr != nil {
			return result, microerror.Mask(clientErr)
		}
		authHeader := args.scheme + " " + args.token
		infoResponse, _, infoErr := apiClient.GetInfo(authHeader, requestIDHeader, infoActivityName, cmdLine)
		if infoErr != nil {
			return result, microerror.Mask(infoErr)
		}

		result.infoResponse = infoResponse
	}

	return result, nil
}
