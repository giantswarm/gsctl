package util

import (
	"github.com/giantswarm/gscliauth/config"
)

func ChooseScheme(endpoint string, token string, chooseBearerScheme bool) string {
	if chooseBearerScheme {
		return config.Config.ChooseSchemeBearer(endpoint, token)
	}
	return config.Config.ChooseScheme(endpoint, token)
}
