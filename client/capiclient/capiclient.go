package capiclient

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/gsctl/client/capiclient/clusters"
	"github.com/giantswarm/microerror"
	"github.com/go-openapi/strfmt"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type Capiclient struct {
	G8sClient *versioned.Clientset
	Clusters  *clusters.Client
}

// New creates a new Capiclient client
func New(kubeconfigPath string, formats strfmt.Registry) *Capiclient {
	// ensure nullable parameters have default
	if formats == nil {
		formats = strfmt.Default
	}

	if kubeconfigPath == "" {
		return nil
	}

	cli := new(Capiclient)
	cli.Clusters = clusters.New(formats)

	var err error
	cli.G8sClient, err = newClientset(kubeconfigPath)
	if err != nil {
		fmt.Println(color.RedString("The connection to the Kubernetes API could not be established."))
		fmt.Println(err)
		os.Exit(1)
	}

	return cli
}

func newClientset(kubeconfigPath string) (*versioned.Clientset, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// create the clientset
	clientset, err := versioned.NewForConfig(clientConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return clientset, nil
}
