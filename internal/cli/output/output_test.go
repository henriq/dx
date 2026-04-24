package output

import (
	"testing"

	"pilot/internal/testutil"

	"github.com/stretchr/testify/assert"
)

// Under `go test`, os.Stdout is a pipe, so IsStdoutTerminal() always returns
// false. These tests therefore exercise only the piped branch of
// PrintPipeable. The TTY branch (which appends a newline when the value does
// not already end with one) is verified by manual smoke test, e.g.
// `pilot secret get FOO` in a terminal vs piped through `xxd`.
func TestPrintPipeable_WhenPiped(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "value without trailing newline is emitted verbatim",
			value: "secret123",
			want:  "secret123",
		},
		{
			name:  "single trailing newline is preserved verbatim",
			value: "secret123\n",
			want:  "secret123\n",
		},
		{
			name:  "multiple trailing newlines are preserved verbatim",
			value: "secret123\n\n\n",
			want:  "secret123\n\n\n",
		},
		{
			name:  "internal newlines are preserved",
			value: "line1\nline2\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "pem-shaped value with trailing newline is unchanged",
			value: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n",
			want:  "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n",
		},
		{
			name:  "empty value produces no output",
			value: "",
			want:  "",
		},
		{
			name:  "only newlines are preserved",
			value: "\n\n",
			want:  "\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := testutil.CaptureStdout(t, func() {
				PrintPipeable(tc.value)
			})
			assert.Equal(t, tc.want, got)
		})
	}
}
