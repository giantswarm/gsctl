package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// makeMockServer returns a mock server to be used in several test cases
func makeMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("mockServer request: %s %s\n", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && r.URL.String() == "/v4/clusters/test-cluster-id/" {
			// return cluster details
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test-cluster-id",
				"name": "Name of the cluster",
				"api_endpoint": "https://api.foo.bar",
				"create_date": "2017-11-20T12:00:00.000000Z",
				"owner": "acmeorg",
				"kubernetes_version": "",
        "release_version": "0.3.0",
				"workers": [
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}},
					{"aws": {"instance_type": "m3.large"}, "memory": {"size_gb": 5}, "storage": {"size_gb": 50}, "cpu": {"cores": 2}, "labels": {"foo": "bar"}}
				]
			}`))
		} else if r.Method == "POST" && r.URL.String() == "/v4/clusters/test-cluster-id/key-pairs/" {
			// return new key pair
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
	        "certificate_authority_data": "-----BEGIN CERTIFICATE-----\nMIIDTDCCAjSgAwIBAgIUaqrdpSpB34otU1d9i6fxXzlkShswDQYJKoZIhvcNAQEL\nBQAwHTEbMBkGA1UEAxMSbDguazhzLmdpZ2FudGljLmlvMB4XDTE2MTAxMzEzMDE1\nMloXDTI2MDgyMjEzMDIyMlowHTEbMBkGA1UEAxMSbDguazhzLmdpZ2FudGljLmlv\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAym1c6Yd7mgpCnpIyWgHP\nMcvdSXt/GL026V5O/hDtJ4b8UMBgm0/QSOUoyLVWRayiF9stPajZB2fLGRG8twGg\noYrRA7MF5IlVdauUXuTLVNjCQLDA+oc4IKEUvFaOiXLf0S522HNgvAE0BoTHoBdD\n5jHgBNJfiu0G4Aju46nPRqVy2Pq6lh6LkGiPfOod1JqR6+43RQlkTNFczXO/5gYE\nu5XsTjYlw5RxnKfq4t3u8pcIVuBQjXcaCFmAzyNzrnWCQO90x6/dA/eZRhrM8Isp\nn/SmxyQ9Hfy3yOZ4EiSfOxirajU/dl+PunlFZQqlLslsuVbkQljeQDaJj/x6dTPM\n4wIDAQABo4GDMIGAMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0G\nA1UdDgQWBBQVsYaAsm0uUKhxfmRyvZ90SeOEBDAfBgNVHSMEGDAWgBQVsYaAsm0u\nUKhxfmRyvZ90SeOEBDAdBgNVHREEFjAUghJsOC5rOHMuZ2lnYW50aWMuaW8wDQYJ\nKoZIhvcNAQELBQADggEBAMiAjwJYNoahEEORZ8kS3JTxHMuYEnmQ3+zu25uOeiUy\nW6GLLMVFaBk18C/nzr2Dz3uKTDspLXta19lRzPOddWC01gHKfeZWh5B0oOnD8TqH\nLv7qLKLQRXBcFckMWMzsZZGBuAwXAxwJ7nWhNwQWlS8PoA0Ufy5IEL7Nnprf0MjI\ntP6DCN/08jSFkSqtO5dJUmZ+vb0etIbKzGL9u1Ywo3uUxFelO2m4kfsB3+jK1waZ\nD1cC1bS8aKP0L9Zww7gL79ifnBd588Xwt+UuXhm/BRxmWGAkx3nIGkxgMWnq0eE3\n6Bl7UbbHPLA5aa307GD3vZBx0JPCbiaAor7JoPJZ+oc=\n-----END CERTIFICATE-----",
	        "client_certificate_data": "-----BEGIN CERTIFICATE-----\nMIIDUjCCAjqgAwIBAgIUSLkBzjSPsgjTT4y7Xi/XtryuXJgwDQYJKoZIhvcNAQEL\nBQAwHTEbMBkGA1UEAxMSbDguazhzLmdpZ2FudGljLmlvMB4XDTE3MDQxODIwMTQw\nM1oXDTE3MDQxOTIwMTQzM1owITEfMB0GA1UEAxMWYXBpLmw4Lms4cy5naWdhbnRp\nYy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOC9rq8lPlmuZuy1\nu675q2FX05F0jwI+y7Sqbdp+irJvqEW5te8bKt9YKLt+J82+bxszcFyVaWjbIksN\n82SZWkh2rqYrl7tj3y02VL3RJUi4GHGqMuv2QV5S5hiGvTxwYQbIasqWdwejkL64\nFS44EXHmMgbhiLdwzcaPmidm1mNo7P666PVA13Htc9XvxokXvLiooJNA4YgmQQBt\nLtlOr/9KmSsF4IjKIkjBnf3UwBpxVnxZ3PYDTlQnfpD/TBpdcLwpz775UlQ+ccoa\nTLIiciyPx4RkmzdV4JToi5DbwJq4bV2LkjIKeRVPS4dAfYfJrfKqGDw4E57OCVW7\nsRUOAyMCAwEAAaOBhTCBgjAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIw\nHQYDVR0OBBYEFOnklhWcM+5y6YXPogWSvplJQtwOMB8GA1UdIwQYMBaAFBWxhoCy\nbS5QqHF+ZHK9n3RJ44QEMCEGA1UdEQQaMBiCFmFwaS5sOC5rOHMuZ2lnYW50aWMu\naW8wDQYJKoZIhvcNAQELBQADggEBAKlyGEBAVQ4Ge3zv7sKxgegnu8KCOJ95il4e\nCHwYmWaudZdn4lF7BJP4hnmOt/gHQ/50xh460MoVbDzAttwzSrS1ggkvxPCcauXd\ngbAUBqsNZpkh2n2xSf8nE725rEXxfp7PNmsNBo3MMLCNX+BTUu+B5QOUwe1AvTul\nZi9EWKpF2TE8AWgj6B2vs73vJRFoB2r2M2EWyBRRPUmrZuyI8xBQCSJ/uuBHEKn1\n1+JMhX57hilE9BzHNIb8nfOrsMpgliNL/nXHeHu4/zYUO013+On8+sZIxMe0KRQc\n7mNpBKi6SaTupVv0ESd4lsQLptMFdHWp8IpuReXibo0BW1hq8NE=\n-----END CERTIFICATE-----",
	        "client_key_data": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA4L2uryU+Wa5m7LW7rvmrYVfTkXSPAj7LtKpt2n6Ksm+oRbm1\n7xsq31gou34nzb5vGzNwXJVpaNsiSw3zZJlaSHaupiuXu2PfLTZUvdElSLgYcaoy\n6/ZBXlLmGIa9PHBhBshqypZ3B6OQvrgVLjgRceYyBuGIt3DNxo+aJ2bWY2js/rro\n9UDXce1z1e/GiRe8uKigk0DhiCZBAG0u2U6v/0qZKwXgiMoiSMGd/dTAGnFWfFnc\n9gNOVCd+kP9MGl1wvCnPvvlSVD5xyhpMsiJyLI/HhGSbN1XglOiLkNvAmrhtXYuS\nMgp5FU9Lh0B9h8mt8qoYPDgTns4JVbuxFQ4DIwIDAQABAoIBAAeH05ai1NgEdAZy\ngHt4ejmky74P/crBd+nx3AR6QQOBok3TzzjX3DPnrFW8AHFwdCChNJ6lkwakcR26\ntfElAlVzRJ7kzwzEZ/IH5AcIPwuUv5zvaw1lDwOuG2+u9CBWU6n6hTmMmSh0XqFF\nYdBOqKb8Y6i/Xelnqj2BClVPqNdj2JmD7YxAA7VGFe4c0Mzrt8W8HLRKQ6j0Dd/z\nItB41yYDwunkGNRybOXMCm3Nuj/U5EywylI11KpFMepbKeWUzsqNI8KrmjoccpjZ\nacUxkbUZsOCcC2UX6JLeiCX7d+ozcvWADho9qrZLYvkxY4b2sGBs0R6gl2EUx008\nKFiqDUECgYEA85BRXed05hefjfsRaXubXkCknQCRX/FMCB0Prfg0YKLQso8PKGxz\npA5LsW21ondQQX1yWK3rC3qe7n2X8oZpPmo9A+qQhvrNmKMH1U8t+z08OPeoRHkz\nyt58OTRFMirWBVn6MVVoAqQt/dzSHiCGsBk6XZJEReHfCQWn2zABAGkCgYEA7DdT\nvHYQ9d2XqpyHwgmx6blPvKXsmP7fBZgsOC6MmSukOeJ9oTLOpHBxPBXunTjpk6U/\nS/m8/zlf+QvRK8sg6wzrmBehlCA9ePFyOCG/5olGvHIMitX+6vSLrbJusxSQM61q\nuLmwVzREINVJ6slkpGwuu6lJXpAmclJ41rtFNasCgYA0NGesR/L/amrRhNHbmRnZ\nHuPpnviJ5u9UAd6dfEjFucAftZgbIvu6WzIQKqK22voBv4Clz0lE4Zh1J8hMvFCM\nhzrivwERXWp539/K8bi6VAq3byXK32uhfQSFQlXehd3vsbR1pIexoT0WX6FNwcz8\nq7ud2L73d41VoreyvFxKmQKBgQCUaycH8T8y3KqhHn0GZEUPT8pUBAUnFG1Y/IY8\nPrNEwnELlc3N7Th9hdEAKd+llc7dYCTnPeGMk6ZDuzMQSy9BwPp+s8poYeF+Dmbv\n8fS7i2GQojBTQ6ZKRqFE4CpCBxecAMhfjPzJriNoZdtt1GCSFw8+Bl39NqGRj1Qx\nx7TyxQKBgACNkMK3r2epSZ+u1B6VR31QuaPTtTafivQ/htKsrTcQDH7L+UkYXNQY\nQFQczPcw54XOdRqFw0LYLkV1Y4PbVOFrBSb2kKwVvknfXx9TV8pI+W0fKw2duFeM\nj7TT0W7vP42Y4yNqs4hTnKioWj9vLOyh2le/waSNABujfNwh8owK\n-----END RSA PRIVATE KEY-----",
	        "create_date": "2017-04-18T20:17:14.544037411Z",
	        "description": "test",
	        "id": "48:b9:01:ce:34:8f:b2:08:d3:4f:8c:bb:5e:2f:d7:b6:bc:ae:5c:98",
	        "ttl_hours": 24
	      }`))
		}
	}))
}

// Test_CreateKubeconfig tests the createKubeconfig function with expected settings
func Test_CreateKubeconfig(t *testing.T) {
	mockServer := makeMockServer()
	defer mockServer.Close()

	// temporary kubeconfig file
	kubeConfigPath, err := tempKubeconfig()
	if err != nil {
		t.Error(err)
	}
	os.Setenv("KUBECONFIG", kubeConfigPath)
	defer os.Unsetenv("KUBECONFIG")

	configDir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(configDir)

	args := createKubeconfigArguments{
		authToken:   "auth-token",
		apiEndpoint: mockServer.URL,
		clusterID:   "test-cluster-id",
		contextName: "giantswarm-test-cluster-id",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err = verifyCreateKubeconfigPreconditions(args, []string{})
	if err != nil {
		t.Error(err)
	}

	result, err := createKubeconfig(args)
	if err != nil {
		t.Error(err)
	}

	// check result object contents
	if result.apiEndpoint == "" {
		t.Error("Expected non-empty result.apiEndpoint, got empty string")
	}
	if result.caCertPath == "" {
		t.Error("Expected non-empty result.caCertPath, got empty string")
	}
	if result.clientKeyPath == "" {
		t.Error("Expected non-empty result.clientKeyPath, got empty string")
	}
	if result.clientCertPath == "" {
		t.Error("Expected non-empty result.clientCertPath, got empty string")
	}

	// check kubeconfig content
	content, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(content), "current-context: giantswarm-"+cmdClusterID) {
		t.Error("Kubeconfig doesn't contain the expected current-context value")
	}
	if !strings.Contains(string(content), "client-certificate: "+configDir) {
		t.Error("Kubeconfig doesn't contain the expected client-certificate value")
	}
	if !strings.Contains(string(content), "client-key: "+configDir) {
		t.Error("Kubeconfig doesn't contain the expected client-key value")
	}
	if !strings.Contains(string(content), "certificate-authority: "+configDir) {
		t.Error("Kubeconfig doesn't contain the expected certificate-authority value")
	}
}

// Test_CreateKubeconfigSelfContained tests creation of a self-contained
// kubeconfig file with inline certs/key
func Test_CreateKubeconfigSelfContained(t *testing.T) {
	mockServer := makeMockServer()
	defer mockServer.Close()

	// temporary config
	configDir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(configDir)

	// output folder
	tmpdir := tempDir()
	defer os.RemoveAll(tmpdir)

	args := createKubeconfigArguments{
		apiEndpoint:       mockServer.URL,
		authToken:         "auth-token",
		clusterID:         "test-cluster-id",
		contextName:       "giantswarm-test-cluster-id",
		selfContainedPath: tmpdir + string(os.PathSeparator) + "kubeconfig",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err = verifyCreateKubeconfigPreconditions(args, []string{})
	if err != nil {
		t.Error(err)
	}

	result, err := createKubeconfig(args)
	if err != nil {
		t.Error(err)
	}

	// check result object contents
	if result.apiEndpoint == "" {
		t.Error("Expected non-empty result.apiEndpoint, got empty string")
	}
	if result.caCertPath != "" {
		t.Error("Expected empty result.caCertPath, got non-empty string " + result.caCertPath)
	}
	if result.clientKeyPath != "" {
		t.Error("Expected empty result.clientKeyPath, got non-empty string " + result.clientKeyPath)
	}
	if result.clientCertPath != "" {
		t.Error("Expected empty result.clientCertPath, got non-empty string " + result.clientCertPath)
	}
	if result.selfContainedPath == "" {
		t.Error("Expected non-empty result.selfContainedPath, got empty string")
	}

	// check kubeconfig content
	content, err := ioutil.ReadFile(result.selfContainedPath)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(content), "current-context: giantswarm-"+cmdClusterID) {
		t.Error("Kubeconfig doesn't contain the expected current-context value")
	}
	if !strings.Contains(string(content), "client-certificate-data:") {
		t.Error("Kubeconfig doesn't contain the key client-certificate-data")
	}
	if !strings.Contains(string(content), "client-key-data:") {
		t.Error("Kubeconfig doesn't contain the key client-key-data")
	}
	if !strings.Contains(string(content), "certificate-authority-data:") {
		t.Error("Kubeconfig doesn't contain the key certificate-authority-data")
	}
}

// Test_CreateKubeconfigCustomContext tests creation of a kubeconfig
// with custom context name
func Test_CreateKubeconfigCustomContext(t *testing.T) {
	mockServer := makeMockServer()
	defer mockServer.Close()

	// temporary kubeconfig file
	kubeConfigPath, err := tempKubeconfig()
	if err != nil {
		t.Error(err)
	}
	os.Setenv("KUBECONFIG", kubeConfigPath)
	defer os.Unsetenv("KUBECONFIG")

	// temporary config
	configDir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(configDir)

	args := createKubeconfigArguments{
		apiEndpoint: mockServer.URL,
		authToken:   "auth-token",
		clusterID:   "test-cluster-id",
		contextName: "test-context",
	}

	cmdAPIEndpoint = mockServer.URL
	initClient()

	err = verifyCreateKubeconfigPreconditions(args, []string{})
	if err != nil {
		t.Error(err)
	}

	result, err := createKubeconfig(args)
	if err != nil {
		t.Error(err)
	}

	// check result object
	if result.contextName == "" {
		t.Error("Expected non-empty result.contextName, got empty string")
	}

	// check kubeconfig content
	content, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(content), "current-context: "+args.contextName) {
		t.Error("Kubeconfig doesn't contain the expected context name")
	}
}

// Test_CreateKubeconfigNoConnection tests what happens if there is no API connection
func Test_CreateKubeconfigNoConnection(t *testing.T) {
	// temporary kubeconfig file
	kubeConfigPath, err := tempKubeconfig()
	if err != nil {
		t.Error(err)
	}
	os.Setenv("KUBECONFIG", kubeConfigPath)
	defer os.Unsetenv("KUBECONFIG")

	configDir, err := tempConfig("")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(configDir)

	args := createKubeconfigArguments{
		authToken:   "auth-token",
		apiEndpoint: "http://0.0.0.0:12345",
		clusterID:   "test-cluster-id",
		contextName: "giantswarm-test-cluster-id",
	}

	err = verifyCreateKubeconfigPreconditions(args, []string{})
	if err != nil {
		t.Error(err)
	}

	_, err = createKubeconfig(args)
	if err == nil {
		t.Error("Expected error (no connection, no response) didn't occur.")
	}

}
