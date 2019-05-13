package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", ProgramName)
	if err != nil {
		panic(err)
	}
	return dir
}

// tempConfig creates a temporary config directory with config.yaml file
// containing the given YAML content and initializes our config from it.
// The directory path ist returned.
func tempConfig(configYAML string) (string, error) {
	dir := tempDir()
	filePath := path.Join(dir, ConfigFileName+"."+ConfigFileType)

	if configYAML != "" {
		file, fileErr := os.Create(filePath)
		if fileErr != nil {
			return dir, fileErr
		}
		file.WriteString(configYAML)
		file.Close()
	}

	err := Initialize(dir)
	if err != nil {
		return dir, err
	}

	return dir, nil
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
	testRefreshToken := "some-refresh-token"
	testScheme := "some-scheme"
	testToken := "some-token"
	testAlias := "testalias"

	// check initial enpoint count
	if Config.NumEndpoints() != 0 {
		t.Error("Expected zero endpoints, got", Config.NumEndpoints())
	}

	// directly set some configuration
	Config.LastVersionCheck = time.Time{}
	err = Config.StoreEndpointAuth(testEndpointURL, testAlias, testEmail, testScheme, testToken, testRefreshToken)
	if err != nil {
		t.Error(err)
	}

	err = Config.SelectEndpoint(testEndpointURL)
	if err != nil {
		t.Error(err)
	}

	if Config.NumEndpoints() != 1 {
		t.Error("Expected 1 endpoint, got", Config.NumEndpoints())
	}

	if !Config.HasEndpointAlias(testAlias) {
		t.Errorf("Expected to have alias '%s', but haven't", testAlias)
	}

	u, err := Config.EndpointByAlias(testAlias)
	if err != nil {
		t.Error(err)
	}
	if u != testEndpointURL {
		t.Errorf("Expected to get URL '%s', but got '%s'", testEndpointURL, u)
	}

	err = WriteToFile()
	if err != nil {
		t.Error(err)
	}
	content, readErr := ioutil.ReadFile(ConfigFilePath)
	if readErr != nil {
		t.Error(readErr)
	}
	yamlText := string(content)

	// check if the last updated date has been added
	if !strings.Contains(yamlText, "updated:") {
		t.Log(yamlText)
		t.Error("Written YAML doesn't contain the expected string 'updated:'")
	}

	// check if the selected endpoint is set as expected
	if !strings.Contains(yamlText, "selected_endpoint: "+testEndpointURL) {
		t.Log(yamlText)
		t.Errorf("Written YAML doesn't contain the expected string 'selected_endpoint: %s'", testEndpointURL)
	}

	// check if the alias has been set
	if !strings.Contains(yamlText, "alias: "+testAlias) {
		t.Log(yamlText)
		t.Errorf("Written YAML doesn't contain the expected string 'alias: %s'", testAlias)
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

	if !strings.Contains(yamlText, "selected_endpoint: "+testEndpointURL) {
		t.Log(yamlText)
		t.Errorf("Written YAML doesn't contain the expected string 'selected_endpoint: %s'", testEndpointURL)
	}
}

// Test_Initialize_NonEmpty tests initializing with a dummy config file.
// The config file has one endpoint, which is also the selected one.
func Test_Initialize_NonEmpty(t *testing.T) {
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

	dir, err := tempConfig(yamlText)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	if Config.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, Config.Email)
	}
	if Config.Token != "some-token" {
		t.Errorf("Expected token '%s', got '%s'", token, Config.Token)
	}

	// test what happens after logout
	Config.Logout("https://myapi.domain.tld")

	content, readErr := ioutil.ReadFile(ConfigFilePath)
	if readErr != nil {
		t.Error(readErr)
	}
	yamlText = string(content)

	if strings.Contains(yamlText, "some-token") {
		t.Log(yamlText)
		t.Errorf("Written YAML still contains token after logout")
	}
	t.Log(yamlText)
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

var normalizeEndpointTests = []struct {
	in  string
	out string
}{
	{"myapi", "https://myapi"},
	{"myapi.com", "https://myapi.com"},
	{"some.api.server/foo/bar", "https://some.api.server"},
	{"http://127.0.0.1:64703", "http://127.0.0.1:64703"},
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

// TestEndpointAlias tests if endpoints can have aliases
// and if they can be used for selecting endpoints
func TestEndpointAlias(t *testing.T) {
	// our test config YAML
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  https://myapi.domain.tld:
    email: email@example.com
    token: some-token
    alias: myalias
  https://other.endpoint:
    email: email@example.com
    token: some-other-token
    alias:
selected_endpoint: https://other.endpoint`

	dir, err := tempConfig(yamlText)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// first, selected endpoint must have empty alias
	if Config.EndpointConfig(Config.SelectedEndpoint).Alias != "" {
		t.Errorf("Expected alias '', got '%s'", Config.EndpointConfig(Config.SelectedEndpoint).Alias)
	}

	err = Config.SelectEndpoint("myalias")
	if err != nil {
		t.Error(err)
	}

	ep := Config.SelectedEndpoint
	if ep != "https://myapi.domain.tld" {
		t.Errorf("Expected endpoint 'https://myapi.domain.tld', got '%s'", ep)
	}

	// after selection, selected endpoint must have alias
	if Config.EndpointConfig(ep).Alias != "myalias" {
		t.Errorf("Expected alias 'myalias', got '%s'", Config.EndpointConfig(Config.SelectedEndpoint).Alias)
	}

	// try selecting using a non-existing alias/URL
	err = Config.SelectEndpoint("non-existing-alias")
	if !IsEndpointNotDefinedError(err) {
		t.Errorf("Expected endpointNotDefinedError, got '%s'", err)
	}

}
