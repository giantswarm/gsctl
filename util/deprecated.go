package util

import "fmt"

func Deprecated(gsctlCmd string, kubectlgsCmd string, docsURL string) string {
	return fmt.Sprintf(`Command "%s" is deprecated, gsctl is being phased out in favor of our 'kubectl gs' plugin.
We recommend you familiarize yourself with the "kubectl gs %s" command as a replacement for this.
For more details see: %s
`, gsctlCmd, kubectlgsCmd, docsURL)
}
