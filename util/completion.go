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

// GenBashCompletionFn generates a bash function
// that can be used for auto-completion of
// command-line commands or flags
func GetBashCompletionFn(fName, fBody string) string {
	return fmt.Sprintf(template, fName, fBody)
}

// RegisterBashCompletionFn adds a custom completion
// function to the root command config
func RegisterBashCompletionFn(command *cobra.Command, fName, fBody string) {
	if !strings.Contains(command.Root().BashCompletionFunction, fName) {
		command.Root().BashCompletionFunction += GetBashCompletionFn(fName, fBody)
	}
}

// SetFlagBashCompletionFn sets a custom completion
// function for a given command-line flag
func SetFlagBashCompletionFn(completionFn *BashCompletionFunc) {
	completionFn.Flags.SetAnnotation(completionFn.FlagName, cobra.BashCompCustom, []string{completionFn.FnName})
	RegisterBashCompletionFn(completionFn.Command, completionFn.FnName, completionFn.FnBody)
}

// SetCommandBashCompletion adds functionality
// to the global custom command-line completion function
//
// Due to `cobra` limitations, this is currently the
// only way of adding auto-completion for commands
func SetCommandBashCompletion(completionFn *BashCompletionFunc) {
	customCompletionFn += completionFn.FnBody + "\n"
}

// GetCustomCommandCompletionFnBody returns the
// global custom command-line completion function body
func GetCustomCommandCompletionFnBody() string {
	return customCompletionFn
}
