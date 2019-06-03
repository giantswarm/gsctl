package config

import (
	"path"
	"testing"

	"github.com/spf13/afero"
)

func TestIsCertExpired(t *testing.T) {
	// certificate taken from Go docs. It expired 2014-05-29 00:00:00 +0000 UTC
	expiredCertPEM := `
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

	notYetExpiredCertPEM := `-----BEGIN CERTIFICATE-----
MIIFXTCCA0WgAwIBAgIJAI+hBhVWhRTiMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcxMjA0MTY0MjIzWhcNMjcxMjAyMTY0MjIzWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIIC
CgKCAgEAwsupgMREynsTA/SN65mN+2nb22ExRqtEWcz7EOYI+MmM9dpNpgj1uwDY
I84ekrGNs6nILC1tbSas9VAVuD577Z42CmjCAOJVhjSC/moIuiDgLOcQU058yGoT
v8ljwd4hnc332Rw2/zYmsh/o/TS3RmXTLT0B0qFsz0dJjt89/X1waAnHZvuNgSSC
krq71BRnagjGQHxdu5nvfC0F6LUolzSPQbElmuTd/TExfLBpxbI7UpuhsapdkBs1
JWsqKGN0vclUVWgDgu1xICv+vc3lCjAIdLEqzccD4caSk1LLryqkMpZT0IvfuyRC
nQF4TnkLP/KxThIuJCrdqT4YLN8Z2MxPfz9UjXEIqhoO7mgoHvwukmgcognYA8mC
xg0r4eQjtjChGCCQAWos2XocRK8CeO0lmML39ql7F3py801/KALewvHKqsuRro+S
DXG87RfRrPUbyfZnMG3WFplH2pVx6/I8ogsdF8rxELRbsFzyLShg2nJvmbSYRCZo
jgSLOdcI3jogxZoyhCds64EWDUK24HeaCqQd/HrQErP0mf4JOE06Y2dVekpruhfo
WC+Ea1H5uDhu5REdvEiDBPR/5DRAqRcYIkQVfo3U34K0PU8FqxeyB6Y7cOv5x5uO
n+BdEKNnm4g38cPTLMvmGzVNxMVFh/r4cYQ/6Lm9wNVirgVR5VUCAwEAAaNQME4w
HQYDVR0OBBYEFOj1D9O3LPTX7/LJDH3X70y+18mBMB8GA1UdIwQYMBaAFOj1D9O3
LPTX7/LJDH3X70y+18mBMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIB
AEO6bx8cpwQcC5AgGok2F77yqP6Li9axkY+IA2CwmUfcMR1lgnymcD3HgJRdK64a
n2RXHZrS/HzIezlVT1+DdcXv8XQeM04gXFBlI1OljhW6HZwWTx9AkuBTJjhoIYEJ
G8x/ZqGqBT6PZ6nAp3G/eWS9lBgslXa9RfWrIYH2ovpxTuOpIdD+jx+Y5dlQ8u1n
nEv+faHWDKnQsv7yP+8w4EGsrQwHZSwC3YmJ4brlkNtdRqUY5pyTD0wqreH9skPf
VgNfn45KFTCOcSYc/U7r5ujDRMl7rP2h0suwPHGM0znGzSO2xnFzt9SmodPb1W33
mkf2NIsP2e6fImyXKVOBc/zAjpa5Wq7nmLLnCXXdef9tVdS0RQ9iJjBPt/OUPlG9
VxVn9KIrPU0fC8xdo4JRg/2L5WnwrrR0NJBYaFAahoFVPRW8zufv864US+yDFkxQ
h17hAY+JzCA/cQ45K6m5mSr+OKtVXoIZXM15xXZ+CgwvrZ+fexrJcqKNxwmpB9nn
XyMBgG8cJVFvjbj/tRgws7eSZd6kmw3WEzuNuYFwALqY089ldCP2PfvjjXwCQgGa
N4K874n07iiCmajfeQxHk3Zle9noUzv0HCrr42H7P1MSwurIBm8UlNdCruKHapWs
WGnPiXqCuccNAHWN9e5ULL3WoKfoLdshSyA9aQ44F3nJ
-----END CERTIFICATE-----`

	expired, err := isCertExpired([]byte(expiredCertPEM))
	if err != nil {
		t.Error(err)
	}
	if expired != true {
		t.Error("Expected true, got false")
	}

	expired, err = isCertExpired([]byte(notYetExpiredCertPEM))
	if err != nil {
		t.Error(err)
	}
	if expired == true {
		t.Error("Expected false, got true")
	}
}

func TestGarbageCollectKeyPairs(t *testing.T) {
	// temporary config dir
	fs := afero.NewMemMapFs()
	_, tempConfigErr := tempConfig(fs, "")
	if tempConfigErr != nil {
		t.Error(tempConfigErr)
	}

	// copy test files over to temporary certs dir
	basePath := "testdata"
	testdataFs := afero.NewOsFs()
	files, _ := afero.ReadDir(testdataFs, basePath)
	for _, f := range files {
		originPath := basePath + "/" + f.Name()
		targetPath := CertsDirPath + "/" + f.Name()
		t.Logf("Copying %s to %s", originPath, targetPath)

		from, oerr := testdataFs.Open(originPath)
		if oerr != nil {
			t.Error(oerr)
		}

		err := afero.WriteReader(fs, targetPath, from)
		if err != nil {
			t.Error(err)
		}
	}

	err := GarbageCollectKeyPairs(fs)
	if err != nil {
		t.Error(err)
	}

	// Check remaining files.
	// test1 should be removed
	exists, err := afero.Exists(fs, path.Join(CertsDirPath, "/test1-client.crt"))
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("test1-client.crt should have been deleted, is still there")
	}

	exists, err = afero.Exists(fs, path.Join(CertsDirPath, "/test1-client.key"))
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("test1-client.key should have been deleted, is still there")
	}

	// test2 should still exist
	exists, err = afero.Exists(fs, path.Join(CertsDirPath, "/test2-client.crt"))
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("test2-client.crt should have been kept, was deleted")
	}

	exists, err = afero.Exists(fs, path.Join(CertsDirPath, "/test2-client.key"))
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("test2-client.key should have been kept, was deleted")
	}

	files, err = afero.ReadDir(fs, CertsDirPath)
	if err != nil {
		t.Error(err)
	}

	for _, file := range files {
		t.Log(file.Name())
	}

}
