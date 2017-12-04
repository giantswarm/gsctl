package config

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
)

// GarbageCollectKeyPairs removes files from expired key pairs
func GarbageCollectKeyPairs() error {
	files, err := ioutil.ReadDir(CertsDirPath)
	if err != nil {
		return microerror.Maskf(err, "could not list files in certs folder"+CertsDirPath)
	}

	// find out which certificates in certs folder have expired
	expiredCerts := []string{}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, "-client.crt") {
			path := CertsDirPath + "/" + name

			// read file content
			content, err := ioutil.ReadFile(path)
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

	for _, file := range expiredCerts {
		fmt.Printf("Certificate %s is expired and will be deleted.\n", file)
		certPath := CertsDirPath + "/" + file
		err := os.Remove(certPath)
		if err != nil {
			log.Printf("Certificate %s could not be deleted.", certPath)
		}

		keyPath := CertsDirPath + "/" + strings.Replace(file, ".crt", ".key", 1)
		err = os.Remove(keyPath)
		if err != nil {
			log.Printf("Key %s could not be deleted.", keyPath)
		}
	}

	return nil
}

func isCertExpired(pemContent []byte) (bool, error) {
	expired := false

	block, _ := pem.Decode(pemContent)
	if block == nil {
		return expired, microerror.Mask(errors.New("could not parse PEM"))
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return expired, microerror.Maskf(errors.New("could not parse certificate"), err.Error())
	}

	if cert.NotAfter.Before(time.Now()) {
		expired = true
	}

	return expired, nil
}
