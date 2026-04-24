package testutil

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// CaptureStdout runs fn with os.Stdout redirected to a pipe and returns
// everything written during the call. The original os.Stdout is restored via
// t.Cleanup.
//
// Tests using this helper must NOT call t.Parallel(): os.Stdout is process
// global, so concurrent swaps collide.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("capture stdout: failed to create pipe: %v", err)
	}
	os.Stdout = writer

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, reader)
		done <- buf.Bytes()
	}()

	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("capture stdout: failed to close writer: %v", err)
	}
	captured := <-done
	if err := reader.Close(); err != nil {
		t.Fatalf("capture stdout: failed to close reader: %v", err)
	}
	return string(captured)
}
