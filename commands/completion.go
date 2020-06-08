package commands

// The 'completion' command is defined on the top leve of the commands
// package, as it has to have access to the root command.

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// CompletionCommand is the command to create things
	CompletionCommand = &cobra.Command{
		Use:                   "completion <bash|fish|zsh> [--stdout]",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Short:                 "Create completion file for bash, fish, or zsh",
		Long: `Generates shell completion code to tab-complete gsctl's commands and
some required flags.

The completion file will either be written to the current directory
or, when adding the --stdout flag, be written to the standard output.

Bash
----

To enable bash completion for gsctl, add a line like this
to your ~/.bash_profile

    source <(gsctl completion bash --stdout)

Then start a new terminal session.

Fish
----

Run

    gsctl completion fish --stdout > ~/.config/fish/completions/gsctl.fish

and then start a new terminal session.

Zsh
---

The generic installation procedure for zsh is:

1. Decide for a target directory from your $FPATH, e. g. ~/.my_completion
   and execute this:

   gsctl completion zsh --stdout > ~/.my_completion/_gsctl

2. Re-initialize your completion files:

   rm -f ~/.zcompdump; compinit
`,
		PreRun: validateCompletionPreconditions,
		Run:    generateCompletionFile,
	}

	// cmdStdOut is the --stdout flag.
	cmdStdOut bool
)

const (
	completionFileNameBash string = "gsctl-completion-bash.sh"
	completionFileNameFish string = "gsctl-completion-fish.fish"
	completionFileNameZsh  string = "gsctl-completion-zsh.sh"

	shellBash = "bash"
	shellFish = "fish"
	shellZsh  = "zsh"
)

func init() {
	CompletionCommand.Flags().BoolVarP(&cmdStdOut, "stdout", "", false, "Write to standard output instead of a file.")
}

// validateCompletionPreconditions validates user input.
func validateCompletionPreconditions(cmd *cobra.Command, args []string) {
	shell := args[0]
	if shell != shellBash && shell != shellFish && shell != shellZsh {
		fmt.Println(color.RedString("Shell not supported"))
		fmt.Println("We only provide shell completion support for bash, fish, and zsh at this time.")
		os.Exit(1)
	}
}

// generateCompletionFile creates bash or zsh completion files.
func generateCompletionFile(cmd *cobra.Command, args []string) {
	shell := args[0]

	switch shell {
	case shellBash:
		if cmdStdOut {
			RootCommand.GenBashCompletion(os.Stdout)
		} else {
			RootCommand.GenBashCompletionFile(completionFileNameBash)
			fmt.Printf("Created completion file for %s in %s\n", shell, completionFileNameBash)
			os.Chmod(completionFileNameBash, 0777)
		}
	case shellFish:
		if cmdStdOut {
			RootCommand.GenFishCompletion(os.Stdout, true)
		} else {
			RootCommand.GenFishCompletionFile(completionFileNameFish, true)
			fmt.Printf("Created completion file for %s in %s\n", shell, completionFileNameFish)
			os.Chmod(completionFileNameFish, 0777)
		}
	case shellZsh:
		if cmdStdOut {
			RootCommand.GenZshCompletion(os.Stdout)
		} else {
			RootCommand.GenZshCompletionFile(completionFileNameZsh)
			fmt.Printf("Created completion file for %s in %s\n", shell, completionFileNameZsh)
			os.Chmod(completionFileNameZsh, 0777)
		}
	}
}
