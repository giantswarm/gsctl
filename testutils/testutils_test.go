package testutils

import (
	"fmt"
	"testing"
)

func TestCaptureOutput(t *testing.T) {
	input := "This is the first line\n"
	f := func() {
		fmt.Printf(input)
	}

	output := CaptureOutput(f)
	if output != input {
		t.Errorf("Expected %q, got %q", input, output)
	}
}

func TestCaptureOutputSync(t *testing.T) {
	input := "This is the first line\n"
	f := func() {
		fmt.Printf(input)
	}

	output := CaptureOutputSync(f)
	if output != input {
		t.Errorf("Expected %q, got %q", input, output)
	}
}
