package util

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/testutils"
	"github.com/spf13/afero"
)

func TestGetClusterID(t *testing.T) {
	var testCases = []struct {
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
			id, err := GetClusterID(tc.clusterNameOrID, clientWrapper)

			if id != tc.expectedID {
				t.Errorf("Case %d - Result did not match ", i)
			} else if err == nil && tc.errorMatcher != nil {
				t.Errorf("Case %d - Expected an error but didn't get one. Should I be happy or not? ", i)
			} else if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Errorf("Case %d - Error did not match expected type. Got '%s'", i, err)
			}
		})
	}
}
