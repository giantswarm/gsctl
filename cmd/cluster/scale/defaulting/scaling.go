package defaulting

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/ruleengine"

	"github.com/giantswarm/gsctl/cmd/cluster/scale/request"
)

type ScalingConfig struct {
	AutoScalingEnabled       *bool
	CurrentScalingMax        *int64
	CurrentScalingMin        *int64
	DesiredNumWorkers        *int64
	DesiredNumWorkersChanged *bool
	DesiredScalingMax        *int64
	DesiredScalingMaxChanged *bool
	DesiredScalingMin        *int64
	DesiredScalingMinChanged *bool
}

type Scaling struct {
	autoScalingEnabled       bool
	currentScalingMax        int64
	currentScalingMin        int64
	desiredNumWorkers        int64
	desiredNumWorkersChanged bool
	desiredScalingMax        int64
	desiredScalingMaxChanged bool
	desiredScalingMin        int64
	desiredScalingMinChanged bool
}

func NewScaling(config ScalingConfig) (*Scaling, error) {
	if config.AutoScalingEnabled == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AutoScalingEnabled must not be empty", config)
	}

	s := &Scaling{
		autoScalingEnabled:       *config.AutoScalingEnabled,
		currentScalingMax:        *config.CurrentScalingMax,
		currentScalingMin:        *config.CurrentScalingMin,
		desiredNumWorkers:        *config.DesiredNumWorkers,
		desiredNumWorkersChanged: *config.DesiredNumWorkersChanged,
		desiredScalingMax:        *config.DesiredScalingMax,
		desiredScalingMaxChanged: *config.DesiredScalingMaxChanged,
		desiredScalingMin:        *config.DesiredScalingMin,
		desiredScalingMinChanged: *config.DesiredScalingMinChanged,
	}

	return s, nil
}

func (s *Scaling) Default(ctx context.Context, scaling request.Scaling) request.Scaling {
	rules := []ruleengine.DefaultingRule{
		{
			Conditions: []func() bool{
				func() bool { return s.autoScalingEnabled },
				func() bool { return s.desiredScalingMinChanged },
			},
			Defaulting: func() {
				scaling.Min = s.desiredScalingMin
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return s.autoScalingEnabled },
				func() bool { return s.desiredScalingMaxChanged },
			},
			Defaulting: func() {
				scaling.Max = s.desiredScalingMax
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return s.autoScalingEnabled },
				func() bool { return !s.desiredScalingMinChanged },
			},
			Defaulting: func() {
				scaling.Min = s.currentScalingMin
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return s.autoScalingEnabled },
				func() bool { return !s.desiredScalingMaxChanged },
			},
			Defaulting: func() {
				scaling.Max = s.currentScalingMax
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return s.autoScalingEnabled },
				func() bool { return !s.desiredScalingMinChanged },
				func() bool { return !s.desiredScalingMaxChanged },
				func() bool { return s.desiredNumWorkersChanged },
			},
			Defaulting: func() {
				scaling.Min = s.desiredNumWorkers
				scaling.Max = s.desiredNumWorkers
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return !s.autoScalingEnabled },
				func() bool { return s.desiredScalingMinChanged },
			},
			Defaulting: func() {
				scaling.Min = s.desiredScalingMin
				scaling.Max = s.desiredScalingMin
			},
		},
		{
			Conditions: []func() bool{
				func() bool { return !s.autoScalingEnabled },
				func() bool { return s.desiredScalingMaxChanged },
			},
			Defaulting: func() {
				scaling.Min = s.desiredScalingMax
				scaling.Max = s.desiredScalingMax
			},
		},
	}

	ruleengine.ExecuteDefaulting(ctx, rules)

	return scaling
}
