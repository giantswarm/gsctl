package util

// TODO: not sure how the Kubectl wrapper functions deal with whitespace in arguments.

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/fatih/color"
)

const (
	// name of the kubectl executable
	binaryName string = "kubectl"

	// url to intallatzion instructions
	kubectlInstallURL string = "http://kubernetes.io/docs/user-guide/prereqs/"

	kubectlWindowsInstallURL string = "https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md"
)

// CheckKubectl checks if kubectl exists, is silent if yes, reports error if not
func CheckKubectl() {
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
			return
		}
		// kubectl not installed
		fmt.Println(color.RedString(binaryName + " does not appear to be installed"))
		if runtime.GOOS == "darwin" {
			fmt.Println("Please install via 'brew install kubernetes-cli' or visit")
			fmt.Println(kubectlInstallURL + " for information on how to install " + binaryName + ".")
		} else if runtime.GOOS == "linux" {
			fmt.Println("Please install via 'brew install kubernetes-cli' or visit")
			fmt.Println(kubectlInstallURL + " for information on how to install " + binaryName + ".")
		} else if runtime.GOOS == "windows" {
			fmt.Printf("Please visit %s to download a recent %s binary.\n", kubectlWindowsInstallURL, binaryName)
		}
		os.Exit(1)
	}
}

// KubectlSetCluster is a wrapper for the `kubectl config set-cluster` command
func KubectlSetCluster(clusterID, apiEndpoint, caCertificatePath string) {
	clusterName := "giantswarm-" + clusterID
	serverArgument := "--server=" + apiEndpoint
	certificateAuthorityArgument := "--certificate-authority=" + caCertificatePath
	cmd := exec.Command(binaryName,
		"config", "set-cluster",
		clusterName,
		serverArgument,
		certificateAuthorityArgument)
	if err := cmd.Run(); err != nil {
		fmt.Println(color.RedString("Could not set cluster using 'kubectl config set-cluster ...'"))
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}
}

// KubectlSetCredentials is a wrapper for the `kubectl config set-credentials` command
func KubectlSetCredentials(clusterID, keyPath, certificatePath string) {
	clientKeyArgument := "--client-key=" + keyPath
	clientCertificateArgument := "--client-certificate=" + certificatePath
	userName := "giantswarm-" + clusterID + "-user"
	cmd := exec.Command(binaryName,
		"config", "set-credentials",
		userName,
		clientKeyArgument,
		clientCertificateArgument)
	if err := cmd.Run(); err != nil {
		fmt.Println(color.RedString("Could not set credentials using 'kubectl config set-credentials ...'"))
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}
}

// KubectlSetContext is a wrapper for the `kubectl config set-context` command
func KubectlSetContext(clusterID string) {
	contextName := "giantswarm-" + clusterID
	clusterArgument := "--cluster=giantswarm-" + clusterID
	userArgument := "--user=giantswarm-" + clusterID + "-user"
	cmd := exec.Command(binaryName,
		"config", "set-context",
		contextName,
		clusterArgument,
		userArgument)
	if err := cmd.Run(); err != nil {
		fmt.Println(color.RedString("Could not set context using 'kubectl config set-context ...'"))
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}
}

// KubectlUseContext applies the context for the given cluster ID
func KubectlUseContext(clusterID string) {
	contextName := "giantswarm-" + clusterID
	cmd := exec.Command(binaryName, "config", "use-context", contextName)
	if err := cmd.Run(); err != nil {
		fmt.Println(color.RedString("Could not apply context using 'kubectl config use-context " + contextName + "'"))
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}
}
