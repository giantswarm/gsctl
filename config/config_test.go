package config

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func tempDir() string {
	dir, _ := ioutil.TempDir("", ProgramName)
	return dir
}

// Test_Initialize_Empty tests the case where a config file and its directory
// do not yet exist.
// Configuration is created and then serialized to the YAML file.
// We roughtly check the YAML whether it contains the expected info.
func Test_Initialize_Empty(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	// additional non-existing sub directory
	dir = path.Join(dir, "subdir")

	err := Initialize(dir)
	if err != nil {
		t.Error("Error in Initialize:", err)
	}

	testEndpointURL := "https://myapi.domain.tld"
	testEmail := "user@domain.tld"
	testToken := "some-token"

	// directly set some configuration
	Config.LastVersionCheck = time.Time{}
	Config.SelectEndpoint(testEndpointURL, testEmail, testToken)

	err = WriteToFile()
	if err != nil {
		t.Error(err)
	}
	content, readErr := ioutil.ReadFile(ConfigFilePath)
	if readErr != nil {
		t.Error(readErr)
	}
	yamlText := string(content)

	if !strings.Contains(yamlText, "updated:") {
		t.Log(yamlText)
		t.Error("Written YAML doesn't contain the expected string 'updated:'")
	}

	if !strings.Contains(yamlText, "selected_endpoint: "+testEndpointURL) {
		t.Log(yamlText)
		t.Errorf("Written YAML doesn't contain the expected string 'selected_endpoint: %s'", testEndpointURL)
	}

	// test what happens after logout
	Config.Logout(testEndpointURL)

	err = WriteToFile()
	if err != nil {
		t.Error(err)
	}
	content, readErr = ioutil.ReadFile(ConfigFilePath)
	if readErr != nil {
		t.Error(readErr)
	}
	yamlText = string(content)

	if !strings.Contains(yamlText, `selected_endpoint: ""`) {
		t.Log(yamlText)
		t.Error(`Written YAML doesn't contain the expected string 'selected_endpoint: ""'`)
	}
	t.Log(yamlText)
}

// Test_Initialize_NonEmpty tests initializing with a dummy config file.
// The config file has one endpoint, which is also the selected one.
func Test_Initialize_NonEmpty(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	filePath := path.Join(dir, ConfigFileName+"."+ConfigFileType)

	// write dummy config
	file, fileErr := os.Create(filePath)
	if fileErr != nil {
		t.Error(fileErr)
	}

	// our test config YAML
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  https://myapi.domain.tld:
    email: email@example.com
    token: some-token
selected_endpoint: https://myapi.domain.tld`

	email := "email@example.com"
	token := "some-token"
	file.WriteString(yamlText)
	file.Close()

	err := Initialize(dir)
	if err != nil {
		t.Error("Error in Initialize:", err)
	}

	if Config.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, Config.Email)
	}
	if Config.Token != "some-token" {
		t.Errorf("Expected token '%s', got '%s'", token, Config.Token)
	}
}

// Test_Kubeconfig_Env_Nonexisting tests what happens
// when the KUBECONFIG env variable points to the
// same dir as we use for config, and it's empty
func Test_Kubeconfig_Env_Nonexisting(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	os.Setenv("KUBECONFIG", dir)
	err := Initialize(dir)
	if err != nil {
		t.Error("Error in Initialize:", err)
	}
}

func Test_UserAgent(t *testing.T) {
	ua := UserAgent()
	t.Log(ua)
}

// Test_GetDefaultCluster tests the GetDefaultCluster function
// for the case that only one cluster exists
func Test_GetDefaultCluster(t *testing.T) {
	// returns one cluster
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
      {
        "create_date": "2017-04-16T09:30:31.192170835Z",
        "id": "cluster-id",
        "name": "Some random test cluster",
				"owner": "acme"
      }
    ]`))
	}))
	defer mockServer.Close()

	clusterID, err := GetDefaultCluster("", "", "", mockServer.URL)
	if err != nil {
		t.Error(err)
	}
	if clusterID != "cluster-id" {
		t.Errorf("Expected 'cluster-id', got '%s'", clusterID)
	}
}

var normalizeEndpointTests = []struct {
	in  string
	out string
}{
	{"myapi", "https://myapi"},
	{"myapi.com", "https://myapi.com"},
	{"some.api.server/foo/bar", "https://some.api.server"},
	{"http://localhost:9000", "http://localhost:9000"},
	{"http://localhost:9000/", "http://localhost:9000"},
	{"http://user:pass@localhost:9000/", "http://localhost:9000"},
}

// Test_NormalizeEndpoint tests the normalizeEndpoint function
func Test_NormalizeEndpoint(t *testing.T) {
	for _, tt := range normalizeEndpointTests {
		normalized := normalizeEndpoint(tt.in)
		if normalized != tt.out {
			t.Errorf("normalizeEndpoint('%s') returned '%s', expected '%s'", tt.in, normalized, tt.out)
		}
	}
}
