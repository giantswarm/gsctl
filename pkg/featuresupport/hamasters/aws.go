package hamasters

import (
	"github.com/giantswarm/gsctl/pkg/featuresupport"
)

var (
	aws = featuresupport.Provider{
		Name:            "aws",
		RequiredVersion: "11.5.0",
	}
)
