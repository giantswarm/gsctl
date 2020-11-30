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

func TestInt64Value(t *testing.T) {
	input := int64(3)
	output := Int64Value(input)

	if *output != input {
		t.Errorf("Expected %v, got %v", input, output)
	}
}

func TestInt64ValueZero(t *testing.T) {
	input := int64(0)
	output := Int64Value(input)

	if *output != input {
		t.Errorf("Expected %v, got %v", input, output)
	}
}
