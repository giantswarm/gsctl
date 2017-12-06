package config

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestIsCertExpired(t *testing.T) {
	// certificate taken from Go docs. It expired 2014-05-29 00:00:00 +0000 UTC
	certPEM := `
-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIE31FZVaPXTUwDQYJKoZIhvcNAQEFBQAwSTELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkdvb2dsZSBJbmMxJTAjBgNVBAMTHEdvb2dsZSBJbnRl
cm5ldCBBdXRob3JpdHkgRzIwHhcNMTQwMTI5MTMyNzQzWhcNMTQwNTI5MDAwMDAw
WjBpMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwN
TW91bnRhaW4gVmlldzETMBEGA1UECgwKR29vZ2xlIEluYzEYMBYGA1UEAwwPbWFp
bC5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfRrObuSW5T7q
5CnSEqefEmtH4CCv6+5EckuriNr1CjfVvqzwfAhopXkLrq45EQm8vkmf7W96XJhC
7ZM0dYi1/qOCAU8wggFLMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAa
BgNVHREEEzARgg9tYWlsLmdvb2dsZS5jb20wCwYDVR0PBAQDAgeAMGgGCCsGAQUF
BwEBBFwwWjArBggrBgEFBQcwAoYfaHR0cDovL3BraS5nb29nbGUuY29tL0dJQUcy
LmNydDArBggrBgEFBQcwAYYfaHR0cDovL2NsaWVudHMxLmdvb2dsZS5jb20vb2Nz
cDAdBgNVHQ4EFgQUiJxtimAuTfwb+aUtBn5UYKreKvMwDAYDVR0TAQH/BAIwADAf
BgNVHSMEGDAWgBRK3QYWG7z2aLV29YG2u2IaulqBLzAXBgNVHSAEEDAOMAwGCisG
AQQB1nkCBQEwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL3BraS5nb29nbGUuY29t
L0dJQUcyLmNybDANBgkqhkiG9w0BAQUFAAOCAQEAH6RYHxHdcGpMpFE3oxDoFnP+
gtuBCHan2yE2GRbJ2Cw8Lw0MmuKqHlf9RSeYfd3BXeKkj1qO6TVKwCh+0HdZk283
TZZyzmEOyclm3UGFYe82P/iDFt+CeQ3NpmBg+GoaVCuWAARJN/KfglbLyyYygcQq
0SgeDh8dRKUiaW3HQSoYvTvdTuqzwK4CXsr3b5/dAOY8uMuG/IAR3FgwTbZ1dtoW
RvOTa8hYiU6A475WuZKyEHcwnGYe57u2I2KbMgcKjPniocj4QzgYsVAVKW3IwaOh
yE+vPxsiUkvQHdO2fojCkY8jg70jxM+gu59tPDNbw3Uh/2Ij310FgTHsnGQMyA==
-----END CERTIFICATE-----`

	expired, err := isCertExpired([]byte(certPEM))
	if err != nil {
		t.Error(err)
	}

	if expired != true {
		t.Error("Expected true, got false")
	}
}

func TestGarbageCollectKeyPairs(t *testing.T) {
	// temporary config dir
	dir, tempConfigErr := tempConfig("")
	if tempConfigErr != nil {
		t.Error(tempConfigErr)
	}
	defer os.RemoveAll(dir)

	// copy test files over to temporary certs dir
	basePath := "testdata"
	files, _ := ioutil.ReadDir(basePath)
	for _, f := range files {
		originPath := basePath + "/" + f.Name()
		targetPath := CertsDirPath + "/" + f.Name()
		t.Logf("Copying %s to %s", originPath, targetPath)

		from, oerr := os.Open(originPath)
		if oerr != nil {
			t.Error(oerr)
		}

		to, oerr := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0755)
		if oerr != nil {
			t.Error(oerr)
		}

		_, cerr := io.Copy(to, from)
		if cerr != nil {
			t.Error(cerr)
		}

		to.Close()
		from.Close()
	}

	err := GarbageCollectKeyPairs()
	if err != nil {
		t.Error(err)
	}

	// Check remaining files.
	// test1 should be removed
	if _, err = os.Stat(CertsDirPath + "/test1-client.crt"); err == nil {
		t.Error("test1-client.crt should have been deleted, is still there")
	}
	if _, err = os.Stat(CertsDirPath + "/test1-client.key"); err == nil {
		t.Error("test1-client.key should have been deleted, is still there")
	}
	// test2 should still exist
	if _, err = os.Stat(CertsDirPath + "/test2-client.crt"); err != nil {
		t.Error("test2-client.crt should have been kept, was deleted")
	}
	if _, err = os.Stat(CertsDirPath + "/test2-client.key"); err != nil {
		t.Error("test2-client.key should have been kept, was deleted")
	}

	files, err = ioutil.ReadDir(CertsDirPath)
	if err != nil {
		t.Error(err)
	}

	for _, file := range files {
		t.Log(file.Name())
	}

}
