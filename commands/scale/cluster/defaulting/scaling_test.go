package defaulting

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/gsctl/commands/scale/cluster/request"
)

func Test_Cmd_Cluster_Scale_Defaulting(t *testing.T) {
	testCases := []struct {
		Name                     string
		AutoScalingEnabled       bool
		CurrentScalingMax        int64
		CurrentScalingMin        int64
		DesiredNumWorkers        int64
		DesiredNumWorkersChanged bool
		DesiredScalingMax        int64
		DesiredScalingMaxChanged bool
		DesiredScalingMin        int64
		DesiredScalingMinChanged bool
		ExpectedScaling          request.Scaling
	}{
		{
			Name:                     "case 0 ensures --workers-min=5 sets scaling.min to 5",
			AutoScalingEnabled:       true,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        0,
			DesiredNumWorkersChanged: false,
			DesiredScalingMax:        0,
			DesiredScalingMaxChanged: false,
			DesiredScalingMin:        5,
			DesiredScalingMinChanged: true,
			ExpectedScaling: request.Scaling{
				Min: 5,
				Max: 3,
			},
		},
		{
			Name:                     "case 1 ensures --workers-max=5 sets scaling.max to 5",
			AutoScalingEnabled:       true,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        0,
			DesiredNumWorkersChanged: false,
			DesiredScalingMax:        5,
			DesiredScalingMaxChanged: true,
			DesiredScalingMin:        0,
			DesiredScalingMinChanged: false,
			ExpectedScaling: request.Scaling{
				Min: 3,
				Max: 5,
			},
		},
		{
			Name:                     "case 2 ensures --num-workers=5 sets scaling.max and scaling.min to 5 when AS is enabled",
			AutoScalingEnabled:       true,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        5,
			DesiredNumWorkersChanged: true,
			DesiredScalingMax:        0,
			DesiredScalingMaxChanged: false,
			DesiredScalingMin:        0,
			DesiredScalingMinChanged: false,
			ExpectedScaling: request.Scaling{
				Min: 5,
				Max: 5,
			},
		},
		{
			Name:                     "case 3 ensures --num-workers=5 sets scaling.max and scaling.min to 5 when AS is not enabled",
			AutoScalingEnabled:       false,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        5,
			DesiredNumWorkersChanged: true,
			DesiredScalingMax:        0,
			DesiredScalingMaxChanged: false,
			DesiredScalingMin:        0,
			DesiredScalingMinChanged: false,
			ExpectedScaling: request.Scaling{
				Min: 5,
				Max: 5,
			},
		},
		{
			Name:                     "case 4 ensures --workers-max=5 sets scaling.max and scaling.min to 5 when AS is not enabled",
			AutoScalingEnabled:       false,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        0,
			DesiredNumWorkersChanged: false,
			DesiredScalingMax:        5,
			DesiredScalingMaxChanged: true,
			DesiredScalingMin:        0,
			DesiredScalingMinChanged: false,
			ExpectedScaling: request.Scaling{
				Min: 5,
				Max: 5,
			},
		},
		{
			Name:                     "case 5 ensures --workers-min=5 sets scaling.max and scaling.min to 5 when AS is not enabled",
			AutoScalingEnabled:       false,
			CurrentScalingMax:        3,
			CurrentScalingMin:        3,
			DesiredNumWorkers:        0,
			DesiredNumWorkersChanged: false,
			DesiredScalingMax:        0,
			DesiredScalingMaxChanged: false,
			DesiredScalingMin:        5,
			DesiredScalingMinChanged: true,
			ExpectedScaling: request.Scaling{
				Min: 5,
				Max: 5,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var scaling *Scaling
			{
				c := ScalingConfig{
					AutoScalingEnabled:       &tc.AutoScalingEnabled,
					CurrentScalingMax:        &tc.CurrentScalingMax,
					CurrentScalingMin:        &tc.CurrentScalingMin,
					DesiredNumWorkers:        &tc.DesiredNumWorkers,
					DesiredNumWorkersChanged: &tc.DesiredNumWorkersChanged,
					DesiredScalingMax:        &tc.DesiredScalingMax,
					DesiredScalingMaxChanged: &tc.DesiredScalingMaxChanged,
					DesiredScalingMin:        &tc.DesiredScalingMin,
					DesiredScalingMinChanged: &tc.DesiredScalingMinChanged,
				}

				scaling, err = NewScaling(c)
				if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}
			}

			defaulted := scaling.Default(context.Background(), request.Scaling{})
			if !reflect.DeepEqual(defaulted, tc.ExpectedScaling) {
				t.Fatalf("want matching\n\n%s\n", cmp.Diff(defaulted, tc.ExpectedScaling))
			}
		})
	}
}
