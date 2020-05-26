package featuresupport

import (
	"reflect"
	. "strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_IsSupported(t *testing.T) {
	testCases := []struct {
		providers       []Provider
		providerToCheck string
		versionToCheck  string
		expectedResult  bool
	}{
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			versionToCheck:  "9.0.0",
			expectedResult:  false,
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			versionToCheck:  "10.0.0",
			expectedResult:  true,
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			versionToCheck:  "11.0.0",
			expectedResult:  true,
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "kvm",
			versionToCheck:  "11.0.0",
			expectedResult:  false,
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			versionToCheck:  "not-semver",
			expectedResult:  false,
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "what happened",
				}, {
					Name:            "azure",
					RequiredVersion: "here?",
				},
			},
			providerToCheck: "aws",
			versionToCheck:  "11.0.0",
			expectedResult:  false,
		},
	}

	for i, tc := range testCases {
		t.Run(Itoa(i), func(t *testing.T) {
			feature := Feature{
				Providers: tc.providers,
			}
			result := feature.IsSupported(tc.providerToCheck, tc.versionToCheck)

			if result != tc.expectedResult {
				t.Errorf("Case %d - Expected %t, got %t", i, tc.expectedResult, result)
			}
		})
	}
}

func Test_RequiredVersion(t *testing.T) {
	testCases := []struct {
		providers       []Provider
		providerToCheck string
		expectedResult  *string
	}{
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			expectedResult:  toStringPtr("10.0.0"),
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "azure",
			expectedResult:  toStringPtr("9.0.0"),
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "kvm",
			expectedResult:  nil,
		},
	}

	for i, tc := range testCases {
		t.Run(Itoa(i), func(t *testing.T) {
			feature := Feature{
				Providers: tc.providers,
			}
			result := reflect.ValueOf(feature.RequiredVersion(tc.providerToCheck))
			expectedResult := reflect.ValueOf(tc.expectedResult)

			if result.String() != expectedResult.String() {
				t.Errorf("Case %d - Result did not match", i)
			}
		})
	}
}

func Test_GetProviderWithName(t *testing.T) {
	testCases := []struct {
		providers       []Provider
		providerToCheck string
		expectedResult  *Provider
	}{
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "aws",
			expectedResult: &Provider{
				Name:            "aws",
				RequiredVersion: "10.0.0",
			},
		},
		{
			providers: []Provider{
				{
					Name:            "aws",
					RequiredVersion: "10.0.0",
				}, {
					Name:            "azure",
					RequiredVersion: "9.0.0",
				},
			},
			providerToCheck: "kvm",
			expectedResult:  nil,
		},
	}

	for i, tc := range testCases {
		t.Run(Itoa(i), func(t *testing.T) {
			feature := Feature{
				Providers: tc.providers,
			}
			result := feature.getProviderWithName(tc.providerToCheck)

			if diff := cmp.Diff(tc.expectedResult, result); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}
		})
	}
}

func toStringPtr(str string) *string {
	return &str
}
