package ruleengine

import (
	"context"

	"github.com/giantswarm/microerror"
)

type ValidationRule struct {
	Conditions []func() bool
	Validation func() error
}

func ExecuteValidation(ctx context.Context, rules []ValidationRule) error {
	for _, rule := range rules {
		if rule.Conditions != nil {
			skip := false

			for _, condition := range rule.Conditions {
				if !condition() {
					skip = true
					break
				}
			}

			if skip {
				continue
			}
		}

		err := rule.Validation()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
