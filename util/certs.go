package util

// Utilities for storing certificates/keys to files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fatih/color"

	"github.com/giantswarm/gsctl/config"
)

func writeCredentialFile(fileName, certificateData string) string {
	data := []byte(certificateData)
	err := os.MkdirAll(config.CertsDirPath, 0700)
	if err != nil {
		fmt.Println(color.RedString("Could not create directory", config.CertsDirPath))
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}
	filePath := path.Join(config.CertsDirPath, fileName)
	writeErr := ioutil.WriteFile(filePath, data, 0600)
	if writeErr != nil {
		fmt.Println(color.RedString("Could not create credential file", filePath))
		fmt.Println("Error:")
		fmt.Println(writeErr)
		os.Exit(1)
	}
	return filePath
}

// StoreCaCertificate writes a CA certificate to a file
//
// The file will have the name format `<clusterID>-ca.crt`
func StoreCaCertificate(clusterID, data string) string {
	fileName := clusterID + "-ca.crt"
	return writeCredentialFile(fileName, data)
}

// StoreClientCertificate writes a client certificate to a file
//
// The file will have the name format `<clusterID>-<keypair-id>-client.crt`
func StoreClientCertificate(clusterID, keyPairID, data string) string {
	fileName := clusterID + "-" + CleanKeypairID(keyPairID)[:10] + "-client.crt"
	return writeCredentialFile(fileName, data)
}

// StoreClientKey writes a client key to a file
//
// The file will have the name format `<clusterID>-<keypair-id>-client.key`
func StoreClientKey(clusterID, keyPairID, data string) string {
	fileName := clusterID + "-" + CleanKeypairID(keyPairID)[:10] + "-client.key"
	return writeCredentialFile(fileName, data)
}
