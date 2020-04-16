package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	clientinfo "github.com/giantswarm/gsclientgen/client/info"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/buildinfo"
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

	arguments Arguments
)

// Arguments represents the arguments we can make use of in this command
type Arguments struct {
	apiEndpoint       string
	scheme            string
	token             string
	userProvidedToken string
	verbose           bool
}

// collectArguments returns an Arguments object populated by the user's
// command line arguments and/or config.
func collectArguments() Arguments {
	endpoint := config.Config.ChooseEndpoint(flags.APIEndpoint)
	token := config.Config.ChooseToken(endpoint, flags.Token)
	scheme := config.Config.ChooseScheme(endpoint, flags.Token)

	return Arguments{
		apiEndpoint:       endpoint,
		scheme:            scheme,
		token:             token,
		userProvidedToken: flags.Token,
		verbose:           flags.Verbose,
	}
}

// infoResult is the struct used to return all the info we might want to print
type infoResult struct {
	apiEndpoint          string
	apiEndpointAlias     string
	commitHash           string
	email                string
	token                string
	version              string
	buildDate            string
	configFilePath       string
	kubeConfigPaths      []string
	infoResponse         *clientinfo.GetInfoOK
	environmentVariables map[string]string
}

// validatePreconditions simply returns nil, as the command should work under
// all conditions.
func validatePreconditions(args Arguments) error {
	return nil
}

// printValidation prints if there is anything missing from user input or config.
func printValidation(cmd *cobra.Command, extraArgs []string) {
	arguments = collectArguments()
	err := validatePreconditions(arguments)

	if err != nil {
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)
	}
}

// printInfo prints some information on the current user and configuration.
func printInfo(cmd *cobra.Command, args []string) {
	result, err := info(arguments)

	output := []string{}

	if result.version != buildinfo.VersionPlaceholder && result.version != "" {
		output = append(output, color.YellowString("%s version:", config.ProgramName)+"|"+color.CyanString(result.version)+" - https://github.com/giantswarm/gsctl/releases/tag/"+result.version)
	} else {
		output = append(output, color.YellowString("%s version:", config.ProgramName)+"|"+color.CyanString(buildinfo.VersionPlaceholder))
	}

	if result.buildDate != buildinfo.Placeholder && result.buildDate != "" {
		output = append(output, color.YellowString("%s build:", config.ProgramName)+"|"+color.CyanString(result.buildDate))
	} else {
		output = append(output, color.YellowString("%s build:", config.ProgramName)+"|"+color.RedString(buildinfo.Placeholder))
	}

	if result.commitHash != buildinfo.Placeholder {
		output = append(output, color.YellowString("%s commit hash:", config.ProgramName)+"|"+color.CyanString(result.commitHash)+" - https://github.com/giantswarm/gsctl/commit/"+result.commitHash)
	} else {
		output = append(output, color.YellowString("%s commit hash:", config.ProgramName)+"|"+color.RedString(buildinfo.Placeholder))
	}

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

	if arguments.verbose {
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

	if len(result.environmentVariables) > 0 {
		envTable := []string{}

		for k, v := range result.environmentVariables {
			envTable = append(envTable, color.YellowString(k)+"|"+color.CyanString(v))
		}

		fmt.Printf("\nRelevant environment variables\n")
		fmt.Println(columnize.SimpleFormat(envTable))
	}

	if err != nil {
		fmt.Println()

		// if this is a common error, handle it in the standard way and exit.
		client.HandleErrors(err)
		errors.HandleCommonErrors(err)

		// handle non-standard errors.
		fmt.Println(color.RedString("Some error occurred:"))
		fmt.Println(err.Error())
		os.Exit(1)
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
	result.token = config.Config.ChooseToken(result.apiEndpoint, args.userProvidedToken)
	result.version = buildinfo.Version
	result.buildDate = buildinfo.BuildDate
	result.commitHash = buildinfo.Commit

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
		clientWrapper, err := client.NewWithConfig(args.apiEndpoint, args.userProvidedToken)
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

	result.environmentVariables = getEnvironmentVariables()

	return result, nil
}

func getEnvironmentVariables() map[string]string {
	// all environment variables relevant to gsctl
	vars := []string{
		"GSCTL_CAFILE",
		"GSCTL_CAPATH",
		"GSCTL_DISABLE_CMDLINE_TRACKING",
		"GSCTL_DISABLE_COLORS",
		"GSCTL_ENDPOINT",
		"GSCTL_AUTH_TOKEN",
		"HTTP_PROXY",
		"KUBECONFIG",
	}

	out := make(map[string]string)

	for _, name := range vars {
		val := os.Getenv(name)
		if val != "" {
			out[name] = val
		}
	}

	return out
}
