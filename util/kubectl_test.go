package util

import "testing"

func TestCheckKubectl(t *testing.T) {
	result := CheckKubectl()
	if !result {
		t.Fatal("CheckKubectl() returned false. Kubectl isn't available.")
	}
}
