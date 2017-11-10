package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
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
	token       string
	verbose     bool
	apiEndpoint string
}

// defaultInfoArguments returns an infoArguments object populated by the user's
// command line arguments
func defaultInfoArguments() infoArguments {
	return infoArguments{
		token:       cmdToken,
		verbose:     cmdVerbose,
		apiEndpoint: config.Config.ChooseEndpoint(cmdAPIEndpoint),
	}
}

// infoResult is the struct used to return all the info we might want to print
type infoResult struct {
	apiEndpoint     string
	email           string
	token           string
	version         string
	buildDate       string
	configFilePath  string
	kubeConfigPaths []string
}

func init() {
	RootCommand.AddCommand(InfoCommand)
}

// printInfo prints some information on the current user and configuration
func printInfo(cmd *cobra.Command, args []string) {

	infoArgs := defaultInfoArguments()
	result := info(infoArgs)

	output := []string{}

	if result.apiEndpoint == "" {
		output = append(output, color.YellowString("API endpoint:")+"|n/a")
	} else {
		output = append(output, color.YellowString("API endpoint:")+"|"+color.CyanString(result.apiEndpoint))
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

	output = append(output, color.YellowString("%s version:", config.ProgramName)+"|"+color.CyanString(result.version))
	output = append(output, color.YellowString("%s build:", config.ProgramName)+"|"+color.CyanString(result.buildDate))
	output = append(output, color.YellowString("Config path:")+"|"+color.CyanString(result.configFilePath))

	// kubectl configuration paths
	output = append(output, color.YellowString("kubectl config path:")+"|"+color.CyanString(strings.Join(result.kubeConfigPaths, ", ")))

	fmt.Println(columnize.SimpleFormat(output))
}

// info gets all the information we'd like to show with the "info" command
// and returns it as a struct
func info(args infoArguments) infoResult {
	result := infoResult{}

	if args.apiEndpoint != "" {
		result.apiEndpoint = args.apiEndpoint
	}

	result.email = config.Config.Email
	result.token = config.Config.ChooseToken(result.apiEndpoint, args.token)
	result.version = config.Version
	result.buildDate = config.BuildDate

	result.configFilePath = config.ConfigFilePath

	// kubectl configuration paths
	if len(config.KubeConfigPaths) > 0 {
		for _, myPath := range config.KubeConfigPaths {
			result.kubeConfigPaths = append(result.kubeConfigPaths, myPath)
		}
	}

	return result
}
