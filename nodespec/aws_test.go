package nodespec

import "testing"

func TestAWS(t *testing.T) {
	p, err := NewAWS()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	it, err := p.GetInstanceTypeDetails("p3.2xlarge")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if it.CPUCores != 8 {
		t.Errorf("Expected 8, got %d", it.CPUCores)
	}
	if it.MemorySizeGB != 61 {
		t.Errorf("Expected 61, got %d", it.MemorySizeGB)
	}

}
