package commands

// The 'completion' command is defined on the top leve of the commands
// package, as it has to have access to the root command.

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/flags"
)

var (
	// CompletionCommand is the command to create things
	CompletionCommand = &cobra.Command{
		Use:   "completion",
		Short: "Create completion file for bash, and (experimentally) for zsh",
		Long: `Generates shell completion code to tab-complete gsctl's commands and
some required flags.

The generated completion files will be output in the current folder
with these names:

- gsctl-completion-bash.sh
- gsctl-completion-zsh.sh

Zsh
---

Zsh support is purely experimental (not yet known to work). We would appreciate
your feedback, whether it works for you or not, in this issue:
https://github.com/giantswarm/gsctl/issues/152

The generic installation procedure for zsh is:

1. Rename gsctl-completion-zsh.sh to _gsctl and move it into a directory
   that is part of your $fpath

2. Re-initialize your completion files:

   rm -f ~/.zcompdump; compinit

Bash
----

To enable bash completion for gsctl:

1. Place gsctl-completion-bash.sh somewhere permanently

2. Edit your ~/.bash_profile and add a line like this:

   source /path/to/gsctl-completion-bash.sh

3. Start a new terminal session
`,
		Run: generateCompletionFiles,
	}
)

const (
	completionFileNameBash string = "gsctl-completion-bash.sh"
	completionFileNameZsh  string = "gsctl-completion-zsh.sh"
)

// generateCompletionFiles creates bash and zsh completion files
func generateCompletionFiles(cmd *cobra.Command, args []string) {

	if flags.CmdVerbose {
		fmt.Printf("Creating completion file for bash in %s\n", completionFileNameBash)
	}
	RootCommand.GenBashCompletionFile(completionFileNameBash)
	os.Chmod(completionFileNameBash, 0777)

	if flags.CmdVerbose {
		fmt.Printf("Creating completion file for zsh in %s\n", completionFileNameZsh)
	}
	RootCommand.GenZshCompletionFile(completionFileNameZsh)
	os.Chmod(completionFileNameZsh, 0777)

}
