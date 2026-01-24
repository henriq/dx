//go:build windows

package progress

import (
	"os"

	"golang.org/x/sys/windows"
)

// initTerminal initializes terminal settings for Windows.
// It attempts to enable Virtual Terminal Processing for ANSI escape sequence support.
// Returns true if ANSI sequences are supported, false otherwise.
func initTerminal() bool {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}

	// ENABLE_VIRTUAL_TERMINAL_PROCESSING enables ANSI escape sequence support
	const enableVirtualTerminalProcessing = 0x0004
	if err := windows.SetConsoleMode(handle, mode|enableVirtualTerminalProcessing); err != nil {
		return false
	}

	return true
}
