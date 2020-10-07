package releaseinfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name                       string
		releasesResponse           []byte
		releasesResponseStatusCode int
		infoResponse               []byte
		infoResponseStatusCode     int
		releaseVersion             string
		expectedK8sVersion         string
		expectedK8sVersionEOLDate  *string
		errorMatcher               func(error) bool
	}{
		{
			name: "case 0: getting the information of an existing release",
			releasesResponse: makeReleasesResponse(
				releaseConfig{
					version:    "1.0.0",
					k8sVersion: "15.0.0",
				},
				releaseConfig{
					version:    "1.0.1",
					k8sVersion: "15.0.1",
				},
				releaseConfig{
					version:    "2.0.0",
					k8sVersion: "16.0.0",
				},
			),
			releasesResponseStatusCode: http.StatusOK,
			infoResponse: makeInfoResponse(
				k8sVersionConfig{
					version: "15.0",
					eolDate: "2020-10-20",
				},
				k8sVersionConfig{
					version: "16.0",
					eolDate: "2021-02-01",
				},
			),
			infoResponseStatusCode:    http.StatusOK,
			releaseVersion:            "1.0.1",
			expectedK8sVersion:        "15.0.1",
			expectedK8sVersionEOLDate: toStringPtr("2020-10-20"),
		},
		{
			name: "case 1: getting the information of an existing release, with a k8s version without an EOL date",
			releasesResponse: makeReleasesResponse(
				releaseConfig{
					version:    "1.0.0",
					k8sVersion: "15.0.0",
				},
				releaseConfig{
					version:    "1.0.1",
					k8sVersion: "15.0.1",
				},
				releaseConfig{
					version:    "2.0.0",
					k8sVersion: "16.0.0",
				},
			),
			releasesResponseStatusCode: http.StatusOK,
			infoResponse: makeInfoResponse(
				k8sVersionConfig{
					version: "15.0",
					eolDate: "2020-10-20",
				},
			),
			infoResponseStatusCode:    http.StatusOK,
			releaseVersion:            "2.0.0",
			expectedK8sVersion:        "16.0.0",
			expectedK8sVersionEOLDate: nil,
		},
		{
			name: "case 2: getting the information of a release that doesn't exist",
			releasesResponse: makeReleasesResponse(
				releaseConfig{
					version:    "1.0.0",
					k8sVersion: "15.0.0",
				},
				releaseConfig{
					version:    "1.0.1",
					k8sVersion: "15.0.1",
				},
				releaseConfig{
					version:    "2.0.0",
					k8sVersion: "16.0.0",
				},
			),
			releasesResponseStatusCode: http.StatusOK,
			infoResponse: makeInfoResponse(
				k8sVersionConfig{
					version: "15.0",
					eolDate: "2020-10-20",
				},
			),
			infoResponseStatusCode:    http.StatusOK,
			releaseVersion:            "3.0.0",
			expectedK8sVersion:        "",
			expectedK8sVersionEOLDate: nil,
			errorMatcher:              IsVersionNotFound,
		},
		{
			name:                       "case 3: trying to get the information, but the releases request fails",
			releasesResponse:           []byte{},
			releasesResponseStatusCode: http.StatusUnauthorized,
			infoResponse: makeInfoResponse(
				k8sVersionConfig{
					version: "15.0",
					eolDate: "2020-10-20",
				},
			),
			infoResponseStatusCode:    http.StatusOK,
			releaseVersion:            "3.0.0",
			expectedK8sVersion:        "",
			expectedK8sVersionEOLDate: nil,
			errorMatcher:              IsNotAuthorized,
		},
		{
			name: "case 4: trying to get the information, but the info request fails",
			releasesResponse: makeReleasesResponse(
				releaseConfig{
					version:    "1.0.0",
					k8sVersion: "15.0.0",
				},
				releaseConfig{
					version:    "1.0.1",
					k8sVersion: "15.0.1",
				},
				releaseConfig{
					version:    "2.0.0",
					k8sVersion: "16.0.0",
				},
			),
			releasesResponseStatusCode: http.StatusOK,
			infoResponse:               []byte{},
			infoResponseStatusCode:     http.StatusInternalServerError,
			releaseVersion:             "3.0.0",
			expectedK8sVersion:         "",
			expectedK8sVersionEOLDate:  nil,
			errorMatcher:               IsInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch true {
				case r.Method == http.MethodGet && r.URL.Path == "/v4/releases/":
					w.WriteHeader(tc.releasesResponseStatusCode)
					w.Write(tc.releasesResponse)

				case r.Method == http.MethodGet && r.URL.Path == "/v4/info/":
					w.WriteHeader(tc.infoResponseStatusCode)
					w.Write(tc.infoResponse)

				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer mockServer.Close()

			clientWrapper, err := client.NewWithConfig(mockServer.URL, "")
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			config := Config{
				ClientWrapper:  clientWrapper,
				ReleaseVersion: tc.releaseVersion,
			}
			ri, err := New(config)
			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Fatalf("error not matching expected matcher, got: %s", errors.Cause(err))
				}

				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if tc.expectedK8sVersion != ri.k8sVersion {
				t.Fatalf("k8s version value not expected, got %s", cmp.Diff(tc.expectedK8sVersion, ri.k8sVersion))
			}

			var expectedEolDate *time.Time
			{
				if tc.expectedK8sVersionEOLDate != nil {
					date, _ := time.Parse(strfmt.RFC3339FullDate, *tc.expectedK8sVersionEOLDate)
					expectedEolDate = &date
				}
			}

			diff := cmp.Diff(expectedEolDate, ri.k8sVersionEOLDate)
			if len(diff) > 0 {
				t.Fatalf("k8s version EOL date value not expected, got %s", diff)
			}
		})
	}
}

func TestReleaseInfo_IsK8sVersionEOL(t *testing.T) {
	testCases := []struct {
		name           string
		eolDate        *string
		expectedResult bool
	}{
		{
			name:           "case 0: a date in the past",
			eolDate:        toStringPtr("1975-01-01"),
			expectedResult: true,
		},
		{
			name:           "case 1: a date in the future",
			eolDate:        toStringPtr("2999-01-01"),
			expectedResult: false,
		},
		{
			name:           "case 2: no known EOL date",
			eolDate:        nil,
			expectedResult: false,
		},
		{
			name:           "case 3: the current date",
			eolDate:        toStringPtr(time.Now().String()),
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var eolDate *time.Time
			if tc.eolDate != nil {
				date, _ := time.Parse(strfmt.RFC3339FullDate, *tc.eolDate)
				eolDate = &date
			}
			ri := ReleaseInfo{
				k8sVersionEOLDate: eolDate,
			}

			result := ri.IsK8sVersionEOL()

			if tc.expectedResult != result {
				t.Fatalf("result not expected, got %s", cmp.Diff(tc.expectedResult, result))
			}
		})
	}
}

type releaseConfig struct {
	version    string
	k8sVersion string
}

func makeReleasesResponse(releaseConfigs ...releaseConfig) []byte {
	releases := make([]*models.V4ReleaseListItem, 0, len(releaseConfigs))
	for _, config := range releaseConfigs {
		newRelease := &models.V4ReleaseListItem{
			Version: toStringPtr(config.version),
			Components: []*models.V4ReleaseListItemComponentsItems{
				{
					Name:    toStringPtr("kubernetes"),
					Version: toStringPtr(config.k8sVersion),
				},
			},
		}

		releases = append(releases, newRelease)
	}

	data, _ := json.Marshal(releases)

	return data
}

type k8sVersionConfig struct {
	version string
	eolDate string
}

func makeInfoResponse(k8sVersions ...k8sVersionConfig) []byte {
	response := &models.V4InfoResponse{
		General: &models.V4InfoResponseGeneral{
			KubernetesVersions: make([]*models.V4InfoResponseGeneralKubernetesVersionsItems, 0, len(k8sVersions)),
		},
	}

	for _, config := range k8sVersions {
		date, _ := time.Parse(strfmt.RFC3339FullDate, config.eolDate)
		dateAsStrFmt := strfmt.Date(date)
		newRelease := &models.V4InfoResponseGeneralKubernetesVersionsItems{
			MinorVersion: toStringPtr(config.version),
			EolDate:      &dateAsStrFmt,
		}

		response.General.KubernetesVersions = append(response.General.KubernetesVersions, newRelease)
	}

	data, _ := json.Marshal(response)

	return data
}

func toStringPtr(v string) *string {
	return &v
}
