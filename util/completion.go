package util

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const template = `
%s() {
	%s
}`

type BashCompletionFunc struct {
	Command  *cobra.Command
	Flags    *pflag.FlagSet
	FlagName string
	FnName   string
	FnBody   string
}

func GetBashCompletionFunction(fName, fBody string) string {
	return fmt.Sprintf(template, fName, fBody)
}

func RegisterBashCompletionFunction(command *cobra.Command, fName, fBody string) {
	if !strings.Contains(command.Root().BashCompletionFunction, fName) {
		command.Root().BashCompletionFunction += GetBashCompletionFunction(fName, fBody)
	}
}

func SetBashCompletionFunction(completionFunc BashCompletionFunc) {
	completionFunc.Flags.SetAnnotation(completionFunc.FlagName, cobra.BashCompCustom, []string{completionFunc.FnName})
	RegisterBashCompletionFunction(completionFunc.Command, completionFunc.FnName, completionFunc.FnBody)
}
