package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// CompletionCommand is the command to create things
	CompletionCommand = &cobra.Command{
		Use:   "completion",
		Short: "Create bash and zsh completion code",
		Long: `Lets you create bash and zsh completion code to tab-complete
gsctl's commands.

The generated completion files will be output in the current folder
with these names:

- gsctl-completion-bash.sh
- gsctl-completion-zsh.sh

To enable bash completion:

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

func init() {
	// subcommands
	RootCommand.AddCommand(CompletionCommand)

}

// generateCompletionFiles creates bash and zsh completion files
func generateCompletionFiles(cmd *cobra.Command, args []string) {

	if cmdVerbose {
		fmt.Printf("Creating completion file for bash in %s\n", completionFileNameBash)
	}
	RootCommand.GenBashCompletionFile(completionFileNameBash)
	os.Chmod(completionFileNameBash, 0777)

	if cmdVerbose {
		fmt.Printf("Creating completion file for zsh in %s\n", completionFileNameZsh)
	}
	RootCommand.GenZshCompletionFile(completionFileNameZsh)
	os.Chmod(completionFileNameZsh, 0777)

}
