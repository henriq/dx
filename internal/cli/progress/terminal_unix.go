//go:build !windows

package progress

// initTerminal initializes terminal settings for Unix systems.
// Unix terminals generally support ANSI escape sequences by default.
func initTerminal() bool {
	return true
}
