package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func tempDir() string {
	dir, _ := ioutil.TempDir("", ProgramName)
	return dir
}

// Test_Initialize_Empty tests the case where a config file and its directory
// do not yet exist
func Test_Initialize_Empty(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	// additional non-existing sub directory
	dir = path.Join(dir, "subdir")

	err := Initialize(dir)
	if err != nil {
		t.Error("Error in Initialize:", err)
	}

	Config.Email = "email@example.com"
	Config.Token = "some-token"

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
}

// Test_Initialize_NonEmpty tests initializing with a dummy config file
func Test_Initialize_NonEmpty(t *testing.T) {
	defer viper.Reset()
	dir := tempDir()
	defer os.RemoveAll(dir)
	filePath := path.Join(dir, ConfigFileName+"."+ConfigFileType)

	// write dummy config
	file, fileErr := os.Create(filePath)
	if fileErr != nil {
		t.Error(fileErr)
	}

	email := "email@example.com"
	token := "some-token"
	file.WriteString(fmt.Sprintf("email: %s\ntoken: %s\n", email, token))
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

func Test_Kubeconfig_Env_Nonexisting(t *testing.T) {
	defer viper.Reset()
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
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.String() == "/v1/user/me/memberships" {
			// return organizations for the current user
			w.Write([]byte(`{
          "status_code": 10000,
          "status_text": "success",
          "data": ["org1"]
        }`))
		} else {
			// return clusters for the organization
			w.Write([]byte(`{
          "status_code": 10000,
          "status_text": "success",
          "data": {
            "clusters": [
              {
                "create_date": "2017-04-16T09:30:31.192170835Z",
                "id": "cluster-id",
                "name": "Some random test cluster"
              }
            ]
          }
        }`))
		}
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
