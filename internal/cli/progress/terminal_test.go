package progress

import (
	"strings"
	"testing"
	"time"
)

// Compile-time check that initTerminal exists and returns bool across all platforms
var _ func() bool = initTerminal

func TestClearLine_WithANSISupport(t *testing.T) {
	caps := terminalCapabilities{
		supportsANSI:  true,
		terminalWidth: 80,
	}

	result := clearLine(caps)

	if result != "\033[2K\r" {
		t.Errorf("expected ANSI clear line sequence, got %q", result)
	}
}

func TestClearLine_WithoutANSISupport(t *testing.T) {
	caps := terminalCapabilities{
		supportsANSI:  false,
		terminalWidth: 80,
	}

	result := clearLine(caps)

	// Should start with \r, contain spaces, and end with \r
	if !strings.HasPrefix(result, "\r") {
		t.Error("expected result to start with carriage return")
	}
	if !strings.HasSuffix(result, "\r") {
		t.Error("expected result to end with carriage return")
	}
	// Should have exactly terminalWidth spaces between the \r characters
	inner := result[1 : len(result)-1]
	if len(inner) != 80 {
		t.Errorf("expected 80 spaces, got %d", len(inner))
	}
	if strings.TrimSpace(inner) != "" {
		t.Error("expected inner content to be only spaces")
	}
}

func TestClearLine_WidthVariation(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{40, 40},
		{120, 120},
		{200, 200},
	}

	for _, tc := range tests {
		caps := terminalCapabilities{
			supportsANSI:  false,
			terminalWidth: tc.width,
		}

		result := clearLine(caps)
		inner := result[1 : len(result)-1]

		if len(inner) != tc.expected {
			t.Errorf("width %d: expected %d spaces, got %d", tc.width, tc.expected, len(inner))
		}
	}
}

func TestTruncateToWidth_PlainText(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello\033[0m"},
		{"hello world", 11, "hello world"},
		{"", 10, ""},
		{"test", 0, ""},
	}

	for _, tc := range tests {
		result := truncateToWidth(tc.input, tc.width)
		if result != tc.expected {
			t.Errorf("truncateToWidth(%q, %d) = %q, want %q", tc.input, tc.width, result, tc.expected)
		}
	}
}

func TestTruncateToWidth_WithANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "ANSI codes don't count toward width",
			input:    "\033[1mhello\033[0m",
			width:    5,
			expected: "\033[1mhello\033[0m",
		},
		{
			name:     "truncate with ANSI preserves codes",
			input:    "\033[1mhello world\033[0m",
			width:    5,
			expected: "\033[1mhello\033[0m",
		},
		{
			name:     "multiple ANSI sequences",
			input:    "\033[1mbold\033[0m \033[2mdim\033[0m",
			width:    7,
			expected: "\033[1mbold\033[0m \033[2mdi\033[0m",
		},
		{
			name:     "ANSI at start only",
			input:    "\033[32m+\033[0m test",
			width:    3,
			expected: "\033[32m+\033[0m t\033[0m",
		},
		{
			name:     "only escape codes no visible text",
			input:    "\033[1m\033[0m",
			width:    5,
			expected: "\033[1m\033[0m",
		},
		{
			name:     "long ANSI parameters 256-color",
			input:    "\033[38;2;255;0;0mred\033[0m",
			width:    3,
			expected: "\033[38;2;255;0;0mred\033[0m",
		},
		{
			name:     "truncate 256-color sequence",
			input:    "\033[38;2;255;0;0mred text\033[0m",
			width:    3,
			expected: "\033[38;2;255;0;0mred\033[0m",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := truncateToWidth(tc.input, tc.width)
			if result != tc.expected {
				t.Errorf("got %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestTruncateToWidth_VisibleLength(t *testing.T) {
	// Test that visible length is correctly calculated
	input := "\033[1m\033[32mColored Bold Text\033[0m"
	result := truncateToWidth(input, 10)

	// Count visible characters in result
	visibleCount := 0
	inEscape := false
	for _, r := range result {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		visibleCount++
	}

	if visibleCount != 10 {
		t.Errorf("expected 10 visible characters, got %d", visibleCount)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"zero", 0, "<1s"},
		{"sub-second", 500 * time.Millisecond, "<1s"},
		{"one second", time.Second, "1s"},
		{"five seconds", 5 * time.Second, "5s"},
		{"59 seconds", 59 * time.Second, "59s"},
		{"one minute", 60 * time.Second, "1m 00s"},
		{"one minute one second", 61 * time.Second, "1m 01s"},
		{"one minute thirty seconds", 90 * time.Second, "1m 30s"},
		{"12 minutes 43 seconds", 12*time.Minute + 43*time.Second, "12m 43s"},
		{"one hour", 60 * time.Minute, "60m 00s"},
		{"one hour one minute one second", 61*time.Minute + 1*time.Second, "61m 01s"},
		{"truncates 1.4s to 1s", 1400 * time.Millisecond, "1s"},
		{"truncates 1.9s to 1s", 1900 * time.Millisecond, "1s"},
		{"truncates 2.5s to 2s", 2500 * time.Millisecond, "2s"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDuration(tc.input)
			if result != tc.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
