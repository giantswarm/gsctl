package ruleengine

import "context"

type DefaultingRule struct {
	Conditions []func() bool
	Defaulting func()
}

func ExecuteDefaulting(ctx context.Context, rules []DefaultingRule) {
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

		rule.Defaulting()
	}
}
