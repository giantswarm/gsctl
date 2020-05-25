package hamasters

import (
	"github.com/giantswarm/gsctl/pkg/featuresupport"
)

var (
	HAMasters = featuresupport.Feature{
		Providers: []featuresupport.Provider{
			aws,
		},
	}
)
