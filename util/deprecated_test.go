package util

import (
	"testing"

	"github.com/giantswarm/gsctl/pkg/provider"
)

func TestGetDeprecatedNotice(t *testing.T) {
	var testCases = []struct {
		provider     string
		gsctlCmd     string
		kubectlgsCmd string
		docsURL      string
		output       string
	}{
		{
			provider:     provider.AWS,
			gsctlCmd:     "list something",
			kubectlgsCmd: "get something",
			docsURL:      "https://docs.giantswarm.io/ui-api/kubectl-gs/get-something/",
			output: `Command list something is deprecated, gsctl is being phased out in favor of our kubectl-gs plugin.
We recommend you familiarize yourself with the kubectl gs get something command as a replacement for this.
For more details see: https://docs.giantswarm.io/ui-api/kubectl-gs/get-something/

`,
		},
		{
			provider:     provider.Azure,
			gsctlCmd:     "list something",
			kubectlgsCmd: "get something",
			docsURL:      "https://docs.giantswarm.io/ui-api/kubectl-gs/get-something/",
			output: `Command list something is deprecated, gsctl is being phased out in favor of our kubectl-gs plugin.
We recommend you familiarize yourself with the kubectl gs get something command as a replacement for this.
For more details see: https://docs.giantswarm.io/ui-api/kubectl-gs/get-something/

`,
		},
		{
			provider:     provider.KVM,
			gsctlCmd:     "list anotherthing",
			kubectlgsCmd: "get anotherthing",
			docsURL:      "https://docs.giantswarm.io/ui-api/kubectl-gs/get-anotherthing/",
			output:       "",
		},
	}

	for _, tc := range testCases {
		notice := GetDeprecatedNotice(tc.provider, tc.gsctlCmd, tc.kubectlgsCmd, tc.docsURL)
		if notice != tc.output {
			t.Errorf("Got '%s', wanted '%s'", notice, tc.output)
		}
	}
}
