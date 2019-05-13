package testutils

import (
	"bytes"
	"io"
	"os"
)

// CaptureOutput runs a function and captures returns STDOUT output as a string.
func CaptureOutput(f func()) (printed string) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = orig // restoring the real stdout
	out := <-outC

	return out
}
