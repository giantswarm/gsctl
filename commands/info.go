package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ryanuber/columnize"
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

func init() {
	RootCommand.AddCommand(InfoCommand)
}

// printInfo prints some information on the current user and configuration
func printInfo(cmd *cobra.Command, args []string) {

	output := []string{}

	if config.Config.Organization == "" {
		output = append(output, color.YellowString("Selected organization:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Selected organization:")+"|"+color.CyanString(config.Config.Organization))
	}

	if config.Config.Cluster == "" {
		output = append(output, color.YellowString("Selected cluster:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Selected cluster:")+"|"+color.CyanString(config.Config.Cluster))
	}

	if config.Config.Email == "" {
		output = append(output, color.YellowString("Email:")+"|"+"n/a")
	} else {
		output = append(output, color.YellowString("Email:")+"|"+color.CyanString(config.Config.Email))
	}

	if cmdVerbose {
		if cmdToken != "" {
			output = append(output, color.YellowString("Auth token:")+"|"+color.CyanString(cmdToken))
		} else if config.Config.Token != "" {
			output = append(output, color.YellowString("Auth token:")+"|"+color.CyanString(config.Config.Token))
		} else {
			output = append(output, color.YellowString("Auth token:")+"|n/a")
		}
	}

	output = append(output, color.YellowString("%s version:", config.ProgramName)+"|"+color.CyanString(config.Version))
	output = append(output, color.YellowString("%s build:", config.ProgramName)+"|"+color.CyanString(config.BuildDate))

	output = append(output, color.YellowString("Config path:")+"|"+color.CyanString(config.ConfigFilePath))

	// kubectl configuration paths
	if len(config.KubeConfigPaths) == 0 {
		output = append(output, color.YellowString("kubectl config path:")+"|n/a")
	} else {
		paths := []string{}
		for _, myPath := range config.KubeConfigPaths {
			paths = append(paths, color.CyanString(myPath))
		}
		output = append(output, color.YellowString("kubectl config path:")+"|"+strings.Join(paths, ", "))
	}

	fmt.Println(columnize.SimpleFormat(output))

	config.WriteToFile()
}
