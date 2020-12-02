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

	if vmSize.NumberOfCores != 4 {
		t.Errorf("Expected 4, got %d", vmSize.NumberOfCores)
	}
	if vmSize.MemoryInMB != 17179.869184 {
		t.Errorf("Expected 17179.869184, got %f", vmSize.MemoryInMB)
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
