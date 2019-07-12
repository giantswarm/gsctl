package config

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
)

// GarbageCollectKeyPairs removes files from expired key pairs.
func GarbageCollectKeyPairs(fs afero.Fs) error {
	files, err := afero.ReadDir(fs, CertsDirPath)
	if err != nil {
		return microerror.Maskf(err, "could not list files in certs folder "+CertsDirPath)
	}

	// find out which certificates in certs folder have expired.
	var expiredCerts []string

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, "-client.crt") {
			path := CertsDirPath + "/" + name

			// read file content
			content, err := afero.ReadFile(fs, path)
			if err != nil {
				return microerror.Maskf(err, "could not read file "+path)
			}

			expired, err := isCertExpired(content)
			if err != nil {
				return microerror.Maskf(err, "could not determine if certificate is expired: "+path)
			}

			if expired {
				expiredCerts = append(expiredCerts, name)
			}
		}
	}

	var errorInfo []string

	for _, file := range expiredCerts {
		certPath := CertsDirPath + "/" + file
		err := fs.Remove(certPath)
		if err != nil {
			errorInfo = append(errorInfo, fmt.Sprintf("Certificate file %s could not be deleted (%s)", certPath, err.Error()))
		}

		keyPath := CertsDirPath + "/" + strings.Replace(file, ".crt", ".key", 1)
		err = fs.Remove(keyPath)
		if err != nil {
			errorInfo = append(errorInfo, fmt.Sprintf("Key file %s could not be deleted (%s)", keyPath, err.Error()))
		}
	}

	if len(errorInfo) > 0 {

		if len(expiredCerts)*2 == len(errorInfo) {
			// all deletions failed (2 files per certificate).
			return microerror.Maskf(garbageCollectionFailedError, "%d files not deleted", len(errorInfo))
		}

		// some deletions failed.
		annotation := strings.Join(errorInfo, ", ")
		return microerror.Maskf(garbageCollectionPartiallyFailedError, annotation)
	}

	// success.
	return nil
}

// isCertExpired returns true if the given PEM content represents
// an expired certificate.
func isCertExpired(pemContent []byte) (bool, error) {
	block, _ := pem.Decode(pemContent)
	if block == nil {
		return false, microerror.Mask(errors.New("could not parse PEM"))
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, microerror.Maskf(errors.New("could not parse certificate"), err.Error())
	}

	if cert.NotAfter.After(time.Now()) {
		return false, nil
	}

	return true, nil
}
