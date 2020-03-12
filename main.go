package main

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"

	"github.com/giantswarm/gsctl/commands"
)

func init() {
	columnizeConfig := columnize.DefaultConfig()
	columnizeConfig.Glue = "   "

	// disable color on windows, as it is super slow
	if runtime.GOOS == "windows" || os.Getenv("GSCTL_DISABLE_COLORS") != "" {
		color.NoColor = true
	}

	isKubectl := strings.Contains(os.Args[0], "kubectl-gs")
	os.Setenv("GSCTL_IS_KUBECTL", strconv.FormatBool(isKubectl))
}

func main() {
	commands.RootCommand.Execute()
}
