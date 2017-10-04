package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/giantswarm/gsctl/config"
)

// create a temporary directory
func tempDir() string {
	dir, _ := ioutil.TempDir("", config.ProgramName)
	return dir
}

// create a temporary kubectl config file
func tempKubeconfig() (string, error) {

	// override standard paths for testing
	dir := tempDir()
	config.HomeDirPath = dir
	config.DefaultConfigDirPath = path.Join(config.HomeDirPath, ".config", config.ProgramName)

	// add a test kubectl config file
	kubeConfigPath := path.Join(dir, "tempkubeconfig")
	config.KubeConfigPaths = []string{kubeConfigPath}
	kubeConfig := []byte(`apiVersion: v1
kind: Config
preferences: {}
current-context: g8s-system
clusters:
users:
contexts:
`)
	fileErr := ioutil.WriteFile(kubeConfigPath, kubeConfig, 0700)
	if fileErr != nil {
		return "", fileErr
	}

	return kubeConfigPath, nil
}

func Test_CreateKubeconfig(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("mockServer request: %s %s\n", r.Method, r.URL)
		w.Header().Set("Content-Type", "application/json")
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

	}))
	defer mockServer.Close()

	// temporary kubeconfig file
	kubeConfigPath, err := tempKubeconfig()
	if err != nil {
		t.Error(err)
	}
	os.Setenv("KUBECONFIG", kubeConfigPath)

	configDir, _ := ioutil.TempDir("", config.ProgramName)
	config.Initialize(configDir)
	cmdAPIEndpoint = mockServer.URL
	cmdClusterID = "test-cluster-id"
	checkCreateKubeconfig(CreateKeypairCommand, []string{})
	createKubeconfig(CreateKeypairCommand, []string{})

	// check kubeconfig content
	content, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		t.Error(err)
	}

	// for reference in case of error
	t.Log(string(content))

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
