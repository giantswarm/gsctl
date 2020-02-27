package confirm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/microerror"

	"github.com/fatih/color"
)

// Ask asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func Ask(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", color.YellowString(s))

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(microerror.Mask(err))
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			return false
		}
	}
}

// AskStrict asks the user for confirmation. A user must type the expected confirmation text and
// then press enter.
func AskStrict(s string, c string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s: ", color.YellowString(s))

	for {
		response, err := reader.ReadString('\n')
		response = strings.TrimSuffix(response, "\n")
		if err != nil {
			log.Fatal(microerror.Mask(err))
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case strings.ToLower(c):
			return true
		case "n", "no":
			return false
		default:
			fmt.Printf(color.YellowString("The input entered does not match. "))
			fmt.Printf(color.YellowString("Try again or abort by typing 'n' or 'no': "))
		}
	}
}
