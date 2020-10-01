package nodespec

import (
	"testing"
)

func TestAzure(t *testing.T) {
	p, err := NewAzureProvider()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	vmSize, err := p.GetVMSizeDetails("Standard_D4s_v3")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if vmSize.NumberOfCores != 2 {
		t.Errorf("Expected 2, got %d", vmSize.NumberOfCores)
	}
	if vmSize.MemoryInMB != 8589.934592 {
		t.Errorf("Expected 8589.934592, got %f", vmSize.MemoryInMB)
	}

}

func TestAzureError(t *testing.T) {
	p, err := NewAzureProvider()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	_, err = p.GetVMSizeDetails("non-existing")
	if !IsVMSizeNotFoundErr(err) {
		t.Errorf("Expected 'vm size not found' error, got: %v", err)
	}
}
