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

var customCompletionFn string

func GetBashCompletionFn(fName, fBody string) string {
	return fmt.Sprintf(template, fName, fBody)
}

func RegisterBashCompletionFn(command *cobra.Command, fName, fBody string) {
	if !strings.Contains(command.Root().BashCompletionFunction, fName) {
		command.Root().BashCompletionFunction += GetBashCompletionFn(fName, fBody)
	}
}

func SetFlagBashCompletionFn(completionFn *BashCompletionFunc) {
	completionFn.Flags.SetAnnotation(completionFn.FlagName, cobra.BashCompCustom, []string{completionFn.FnName})
	RegisterBashCompletionFn(completionFn.Command, completionFn.FnName, completionFn.FnBody)
}

func SetCommandBashCompletion(completionFn *BashCompletionFunc) {
	customCompletionFn += completionFn.FnBody + "\n"
}

func GetCustomCommandCompletionFn() string {
	return customCompletionFn
}
