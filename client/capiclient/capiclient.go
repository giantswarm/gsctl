package capiclient

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/gsctl/client/capiclient/clusters"
	"github.com/giantswarm/microerror"
	"github.com/go-openapi/strfmt"
	"k8s.io/client-go/tools/clientcmd"
)

type Capiclient struct {
	Clientset *versioned.Clientset
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

	// Do nothing with the error, because the clientset will be nil
	// if there's an error
	cli.Clientset, _ = newClientset(kubeconfigPath)

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
