package clustercache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func Test_GetClusterID(t *testing.T) {
	testCases := []struct {
		clusterNameOrID string
		expectedID      string
		errorMatcher    func(error) bool
		output          string
	}{
		{
			clusterNameOrID: "My dearest production cluster",
			expectedID:      "fow72",
			errorMatcher:    nil,
		}, {
			clusterNameOrID: "2sg4i",
			expectedID:      "2sg4i",
			errorMatcher:    nil,
		}, {
			clusterNameOrID: "Some cluster that is not here",
			expectedID:      "",
			errorMatcher:    errors.IsClusterNotFoundError,
		}, {
			clusterNameOrID: "A deleted cluster",
			expectedID:      "",
			errorMatcher:    errors.IsClusterNotFoundError,
		},
	}

	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "GET" && r.URL.Path == "/v4/clusters/" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
			{
				"create_date": "2017-05-16T09:30:31.192170835Z",
				"id": "fow72",
				"name": "My dearest production cluster",
				"owner": "acme",
				"path": "/v4/clusters/fow72/"
			},
			{
				"create_date": "2017-04-16T09:30:31.192170835Z",
				"id": "2sg4i",
				"name": "Abandoned cluster from the early days",
				"owner": "some_org",
				"path": "/v4/clusters/2sg4i/"
			},
			{
				"create_date": "2017-10-06T02:24:55.192170835Z",
				"id": "7ste0",
				"name": "A fairly recent test cluster",
				"owner": "acme",
				"path": "/v4/clusters/7ste0/"
			},
			{
				"create_date": "2017-10-20T07:24:55.192170835Z",
				"id": "9as2a",
				"name": "That brand new cluster",
				"owner": "acme_dev",
				"path": "/v4/clusters/9as2a/"
			},{
				"create_date": "2017-10-10T07:24:55.192170835Z",
				"id": "d740d",
				"name": "That brand new cluster",
				"owner": "acme_prod",
				"path": "/v4/clusters/d740d/"
			},
			{
				"create_date": "2017-10-10T07:24:55.192170835Z",
				"delete_date": "2019-10-10T07:24:55.192170835Z",
				"id": "del01",
				"name": "A deleted cluster",
				"owner": "acme",
				"path": "/v5/clusters/del01/"
			}
		]`))
				} else {
					t.Errorf("Case %d - Unsupported operation %s %s called in mock server", i, r.Method, r.URL.Path)
				}
			}))
			defer mockServer.Close()

			// client
			clientWrapper, err := client.NewWithConfig(mockServer.URL, "test-token")
			if err != nil {
				t.Fatalf("Error in client creation: %s", err)
			}

			// output
			id, err := GetID(mockServer.URL, tc.clusterNameOrID, clientWrapper)

			switch {

			case id != tc.expectedID:
				t.Errorf("Case %d - Result did not match ", i)

			case err == nil && tc.errorMatcher != nil:
				t.Errorf("Case %d - Expected an error but didn't get one. Should I be happy or not? ", i)

			case tc.errorMatcher != nil && !tc.errorMatcher(err):
				t.Errorf("Case %d - Error did not match expected type. Got '%s'", i, err)

			}
		})
	}
}

func Test_matchesValidation(t *testing.T) {
	dd := strfmt.NewDateTime()
	deleteDate := &dd

	testCases := []struct {
		clusterNameOrID string
		expectedResult  bool
		cluster         *models.V4ClusterListItem
	}{
		{
			clusterNameOrID: "My dearest production cluster",
			expectedResult:  true,
			cluster: &models.V4ClusterListItem{
				DeleteDate: nil,
				ID:         "123sd",
				Name:       "My dearest production cluster",
			},
		}, {
			clusterNameOrID: "2sg4i",
			expectedResult:  true,
			cluster: &models.V4ClusterListItem{
				DeleteDate: nil,
				ID:         "2sg4i",
				Name:       "Cluster name",
			},
		}, {
			clusterNameOrID: "A deleted cluster",
			expectedResult:  false,
			cluster: &models.V4ClusterListItem{
				DeleteDate: deleteDate,
				ID:         "",
				Name:       "",
			},
		}, {
			clusterNameOrID: "Some other cluster",
			expectedResult:  false,
			cluster: &models.V4ClusterListItem{
				DeleteDate: nil,
				ID:         "123ad12",
				Name:       "Not this cluster",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// output
			isValid := matchesValidation(tc.clusterNameOrID, tc.cluster)

			if isValid != tc.expectedResult {
				t.Errorf("Case %d - Result did not match ", i)
			}
		})
	}
}

func Test_printNameCollisionTable(t *testing.T) {
	testCases := []struct {
		nameOrID       string
		tableLines     []string
		expectedResult string
	}{
		{
			nameOrID: "Cluster name",
			tableLines: []string{
				"ID | ORGANIZATION | RELEASE | CREATED",
				"1asd1 | giantswarm | 11.0.1 | 2020 Mar 04, 09:11 UTC",
				"asd1sd | giantswarm | 11.0.1 | 2020 Mar 04, 09:11 UTC",
			},
			expectedResult: `Multiple clusters found with the name 'Cluster name'.

ID      ORGANIZATION  RELEASE  CREATED
1asd1   giantswarm    11.0.1   2020 Mar 04, 09:11 UTC
asd1sd  giantswarm    11.0.1   2020 Mar 04, 09:11 UTC

`,
		},
	}

	for i, tc := range testCases {
		output := testutils.CaptureOutput(func() {
			// output
			printNameCollisionTable(tc.nameOrID, tc.tableLines)
		})

		if diff := cmp.Diff(tc.expectedResult, output); diff != "" {
			t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
		}
	}
}

func Test_IsInClusterCache(t *testing.T) {
	nonExpiredDate := time.Now().Add(time.Hour * 24).Format(timeLayout)

	testCases := []struct {
		clusterNameOrID string
		cacheYAML       string
		endpoint        string
		expectedResult  bool
	}{
		{
			clusterNameOrID: "My dearest production cluster",
			endpoint:        "mock-endpoint",
			cacheYAML:       "",
			expectedResult:  false,
		}, {
			clusterNameOrID: "2sg4i",
			endpoint:        "mock-endpoint",
			cacheYAML: fmt.Sprintf(`endpoints:
  other-endpoint:
    expiry: "%s"
    ids:
    - 2sg4i`, nonExpiredDate),
			expectedResult: false,
		}, {
			clusterNameOrID: "2sg4i",
			endpoint:        "mock-endpoint",
			cacheYAML: `endpoints:
  mock-endpoint:
    expiry: "2010-03-03T13:17:16+01:00"
    ids:
    - 2sg4i`,
			expectedResult: false,
		}, {
			clusterNameOrID: "2sg4i",
			endpoint:        "mock-endpoint",
			cacheYAML: fmt.Sprintf(`endpoints:
  mock-endpoint:
    expiry: "%s"
    ids:
    - 2sg4i
    - 123asd
    - 1239d1
    - 99sad0`, nonExpiredDate),
			expectedResult: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			_, err := testutils.TempConfig(fs, "")
			_, err = testutils.TempClusterCache(fs, tc.cacheYAML)
			if err != nil {
				t.Fatal(err)
			}

			// output
			isInCache := IsInCache(tc.endpoint, tc.clusterNameOrID)

			if isInCache != tc.expectedResult {
				t.Errorf("Case %d - Result did not match ", i)
			}
		})
	}
}

func Test_CacheClusterIDs(t *testing.T) {
	nonExpiredDate := time.Now().Add(cacheDuration).Format(timeLayout)

	testCases := []struct {
		clusterIDs       []string
		initialCacheYAML string
		cacheYAML        string
		endpoint         string
	}{
		{
			clusterIDs: []string{"1239d1", "99sad0"},
			endpoint:   "other-endpoint",
			initialCacheYAML: `endpoints:
  mock-endpoint:
    expiry: "2010-03-03T13:17:16+01:00"
    ids:
    - 2sg4i
    - 123asd`,
			cacheYAML: fmt.Sprintf(`endpoints:
  mock-endpoint:
    expiry: "2010-03-03T13:17:16+01:00"
    ids:
    - 2sg4i
    - 123asd
  other-endpoint:
    expiry: "%s"
    ids:
    - 1239d1
    - 99sad0
`, nonExpiredDate),
		}, {
			clusterIDs:       []string{"1239d1", "99sad0"},
			endpoint:         "mock-endpoint",
			initialCacheYAML: "",
			cacheYAML: fmt.Sprintf(`endpoints:
  mock-endpoint:
    expiry: "%s"
    ids:
    - 1239d1
    - 99sad0
`, nonExpiredDate),
		}, {
			clusterIDs: []string{"1239d1", "99sad0", "asd10h"},
			endpoint:   "mock-endpoint",
			initialCacheYAML: `endpoints:
  mock-endpoint:
    expiry: "2010-03-03T13:17:16+01:00"
    ids:
    - 1239d1
    - 99sad0
`,
			cacheYAML: fmt.Sprintf(`endpoints:
  mock-endpoint:
    expiry: "%s"
    ids:
    - 1239d1
    - 99sad0
    - asd10h
`, nonExpiredDate),
		}, {
			clusterIDs:       []string{},
			endpoint:         "mock-endpoint",
			initialCacheYAML: "",
			cacheYAML:        "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			_, err := testutils.TempConfig(fs, "")
			clusterCacheFileDir, err := testutils.TempClusterCache(fs, tc.initialCacheYAML)
			if err != nil {
				t.Fatal(err)
			}
			clusterCacheFilePath := path.Join(clusterCacheFileDir, clusterCacheFileName)
			defer fs.Remove(clusterCacheFilePath)

			// output
			CacheIDs(tc.endpoint, tc.clusterIDs)
			cacheContent, _ := afero.ReadFile(fs, clusterCacheFilePath)

			if diff := cmp.Diff(string(cacheContent), tc.cacheYAML); diff != "" {
				t.Errorf("Case %d - Result did not match.\nOutput: %s", i, diff)
			}

		})
	}
}
