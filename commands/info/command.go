package info

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	clientinfo "github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/flags"
)

const (
	infoActivityName = "info"
)

var (
	// Command is the "info" go command
	Command = &cobra.Command{
		Use:    "info",
		Short:  "Print some information",
		Long:   `Prints information that might help you get out of trouble`,
		PreRun: printValidation,
		Run:    printInfo,
	}
)

// Arguments represents the arguments we can make use of in this command
type Arguments struct {
	scheme      string
	token       string
	verbose     bool
	apiEndpoint string
}

// collectArguments returns an Arguments object populated by the user's
// command line arguments and/or config.
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.CmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.CmdToken)
	scheme := config.Config.ChooseScheme(endpoint, flags.CmdToken)

	return Arguments{
		scheme:      scheme,
		token:       token,
		verbose:     flags.CmdVerbose,
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
	infoResponse     *clientinfo.GetInfoOK
}

// validatePreconditions simply returns nil, as the command should work under
// all conditions.
func validatePreconditions(args Arguments) error {
	return nil
}

// printValidation prints if there is anything missing from user input or config.
func printValidation(cmd *cobra.Command, extraArgs []string) {
	args := collectArguments()
	err := validatePreconditions(args)

	if err != nil {
		errors.HandleCommonErrors(err)
	}
}

// printInfo prints some information on the current user and configuration.
func printInfo(cmd *cobra.Command, args []string) {
	infoArgs := collectArguments()
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
		if result.infoResponse == nil || result.infoResponse.Payload.General.Provider == "" {
			output = append(output, color.YellowString("Provider:")+"|n/a")
		} else {
			output = append(output, color.YellowString("Provider:")+"|"+color.CyanString(result.infoResponse.Payload.General.Provider))
		}

		if result.infoResponse != nil {
			if result.infoResponse.Payload.General.Provider == "aws" {
				output = append(output, color.YellowString("Worker EC2 instance type options:")+"|"+color.CyanString(strings.Join(result.infoResponse.Payload.Workers.InstanceType.Options, ", ")))
				output = append(output, color.YellowString("Default worker EC2 instance type:")+"|"+color.CyanString(result.infoResponse.Payload.Workers.InstanceType.Default))
			} else if result.infoResponse.Payload.General.Provider == "azure" {
				output = append(output, color.YellowString("Worker VM size options:")+"|"+color.CyanString(strings.Join(result.infoResponse.Payload.Workers.VMSize.Options, ", ")))
				output = append(output, color.YellowString("Default worker VM size:")+"|"+color.CyanString(result.infoResponse.Payload.Workers.VMSize.Default))
			}

			if result.infoResponse.Payload.Workers.CountPerCluster.Default != 0 {
				output = append(output, color.YellowString("Default workers per cluster:")+"|"+color.CyanString(fmt.Sprintf("%.0f", result.infoResponse.Payload.Workers.CountPerCluster.Default)))
			}
			if result.infoResponse.Payload.Workers.CountPerCluster.Max != 0 {
				output = append(output, color.YellowString("Maximum workers per cluster:")+"|"+color.CyanString(fmt.Sprintf("%.0f", result.infoResponse.Payload.Workers.CountPerCluster.Max)))
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
func info(args Arguments) (infoResult, error) {
	result := infoResult{}

	if args.apiEndpoint != "" {
		result.apiEndpoint = args.apiEndpoint
	}

	result.email = config.Config.Email
	result.token = config.Config.ChooseToken(result.apiEndpoint, args.token)
	result.version = config.Version
	result.buildDate = config.BuildDate

	if config.Config.EndpointConfig(result.apiEndpoint) != nil {
		result.apiEndpointAlias = config.Config.EndpointConfig(result.apiEndpoint).Alias
	}

	result.configFilePath = config.ConfigFilePath

	// kubectl configuration paths
	if len(config.KubeConfigPaths) > 0 {
		for _, myPath := range config.KubeConfigPaths {
			result.kubeConfigPaths = append(result.kubeConfigPaths, myPath)
		}
	}

	// If an endpoint and a token is defined, we pull info from the API, too.
	if args.apiEndpoint != "" && args.token != "" {
		clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.token)
		if err != nil {
			return result, microerror.Mask(err)
		}

		auxParams := clientWrapper.DefaultAuxiliaryParams()
		auxParams.ActivityName = infoActivityName

		response, err := clientWrapper.GetInfo(auxParams)
		if err != nil {
			return result, microerror.Mask(err)
		}

		result.infoResponse = response
	}

	return result, nil
}
