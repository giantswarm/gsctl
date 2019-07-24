package util

// Utilities for storing certificates/keys to files

import (
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/formatting"
)

func writeCredentialFile(fs afero.Fs, certsDirPath, fileName, certificateData string) string {
	data := []byte(certificateData)
	filePath := path.Join(certsDirPath, fileName)
	writeErr := afero.WriteFile(fs, filePath, data, 0600)
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
func StoreCaCertificate(fs afero.Fs, certsDirPath, clusterID, data string) string {
	fileName := clusterID + "-ca.crt"
	return writeCredentialFile(fs, certsDirPath, fileName, data)
}

// StoreClientCertificate writes a client certificate to a file
//
// The file will have the name format `<clusterID>-<keypair-id>-client.crt`
func StoreClientCertificate(fs afero.Fs, certsDirPath, clusterID, keyPairID, data string) string {
	fileName := clusterID + "-" + formatting.CleanKeypairID(keyPairID)[:10] + "-client.crt"
	return writeCredentialFile(fs, certsDirPath, fileName, data)
}

// StoreClientKey writes a client key to a file
//
// The file will have the name format `<clusterID>-<keypair-id>-client.key`
func StoreClientKey(fs afero.Fs, certsDirPath, clusterID, keyPairID, data string) string {
	fileName := clusterID + "-" + formatting.CleanKeypairID(keyPairID)[:10] + "-client.key"
	return writeCredentialFile(fs, certsDirPath, fileName, data)
}
