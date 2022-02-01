package util

import (
	"fmt"

	"github.com/fatih/color"
	p "github.com/giantswarm/gsctl/pkg/provider"
)

func GetDeprecatedNotice(provider, gsctlCmd, kubectlgsCmd, docsURL string) string {
	if provider == p.KVM {
		return ""
	}

	return fmt.Sprintf(`Command %s is deprecated, gsctl is being phased out in favor of our kubectl-gs plugin.
We recommend you familiarize yourself with the %s command as a replacement for this.
For more details see: %s

`, color.YellowString(gsctlCmd), color.YellowString("kubectl gs "+kubectlgsCmd), docsURL)
}
