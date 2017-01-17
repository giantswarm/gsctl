package util

// TODO: not sure how the Kubectl wrapper functions deal with whitespace in arguments.

import (
	"os/exec"
	"syscall"
)

const (
	// name of the kubectl executable
	binaryName string = "kubectl"
)

// CheckKubectl checks if kubectl exists, returns true yes, false otherwise
func CheckKubectl() bool {
	cmd := exec.Command(binaryName)
	if err := cmd.Run(); err != nil {
		// Did the command fail because of an unsuccessful exit code
		var waitStatus syscall.WaitStatus
		exitStatus := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			exitStatus = waitStatus.ExitStatus()
		}
		if exitStatus == 0 {
			return true
		}
		return false
	}
	return true
}

// KubectlSetCluster is a wrapper for the `kubectl config set-cluster` command
func KubectlSetCluster(clusterID, apiEndpoint, caCertificatePath string) error {
	clusterName := "giantswarm-" + clusterID
	serverArgument := "--server=" + apiEndpoint
	certificateAuthorityArgument := "--certificate-authority=" + caCertificatePath
	cmd := exec.Command(binaryName,
		"config", "set-cluster",
		clusterName,
		serverArgument,
		certificateAuthorityArgument)
	err := cmd.Run()
	return err
}

// KubectlSetCredentials is a wrapper for the `kubectl config set-credentials` command
func KubectlSetCredentials(clusterID, keyPath, certificatePath string) error {
	clientKeyArgument := "--client-key=" + keyPath
	clientCertificateArgument := "--client-certificate=" + certificatePath
	userName := "giantswarm-" + clusterID + "-user"
	cmd := exec.Command(binaryName,
		"config", "set-credentials",
		userName,
		clientKeyArgument,
		clientCertificateArgument)
	err := cmd.Run()
	return err
}

// KubectlSetContext is a wrapper for the `kubectl config set-context` command
func KubectlSetContext(clusterID string) error {
	contextName := "giantswarm-" + clusterID
	clusterArgument := "--cluster=giantswarm-" + clusterID
	userArgument := "--user=giantswarm-" + clusterID + "-user"
	cmd := exec.Command(binaryName,
		"config", "set-context",
		contextName,
		clusterArgument,
		userArgument)
	err := cmd.Run()
	return err
}

// KubectlUseContext applies the context for the given cluster ID
func KubectlUseContext(clusterID string) error {
	contextName := "giantswarm-" + clusterID
	cmd := exec.Command(binaryName, "config", "use-context", contextName)
	err := cmd.Run()
	return err
}
