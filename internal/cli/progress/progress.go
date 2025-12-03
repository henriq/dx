package progress

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Status represents the state of a task
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusFailed
)

// Item represents a single task being tracked
type Item struct {
	Name     string
	Info     string // Additional info (e.g., repo, ref)
	Status   Status
	Duration time.Duration
	Error    error
}

// Tracker manages progress display for multiple sequential tasks
type Tracker struct {
	mu           sync.Mutex
	items        []Item
	current      int
	startTime    time.Time
	isTTY        bool
	useColor     bool
	stopChan     chan struct{}
	spinnerFrame int
}

var spinnerFrames = []string{"◐", "◓", "◑", "◒"}

// NewTracker creates a new progress tracker with names only
func NewTracker(names []string) *Tracker {
	items := make([]Item, len(names))
	for i, name := range names {
		items[i] = Item{Name: name, Status: StatusPending}
	}

	_, noColor := os.LookupEnv("NO_COLOR")
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	return &Tracker{
		items:    items,
		current:  -1,
		isTTY:    isTTY,
		useColor: !noColor && isTTY,
		stopChan: make(chan struct{}),
	}
}

// NewTrackerWithInfo creates a new progress tracker with names and additional info
func NewTrackerWithInfo(names []string, infos []string) *Tracker {
	items := make([]Item, len(names))
	for i, name := range names {
		info := ""
		if i < len(infos) {
			info = infos[i]
		}
		items[i] = Item{Name: name, Info: info, Status: StatusPending}
	}

	_, noColor := os.LookupEnv("NO_COLOR")
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	return &Tracker{
		items:    items,
		current:  -1,
		isTTY:    isTTY,
		useColor: !noColor && isTTY,
		stopChan: make(chan struct{}),
	}
}

// Start begins tracking and starts the spinner animation if in TTY mode
func (t *Tracker) Start() {
	if t.isTTY {
		go t.animate()
	}
}

// StartItem marks an item as running
func (t *Tracker) StartItem(index int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current = index
	t.items[index].Status = StatusRunning
	t.startTime = time.Now()

	if !t.isTTY {
		// Non-TTY mode: print timestamped start message
		ts := time.Now().Format("15:04:05")
		item := t.items[index]
		if item.Info != "" {
			fmt.Printf("[%s] Building %s (%s)...\n", ts, item.Name, item.Info)
		} else {
			fmt.Printf("[%s] Building %s...\n", ts, item.Name)
		}
	}
}

// CompleteItem marks an item as completed (success or failure)
func (t *Tracker) CompleteItem(index int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.items[index].Duration = time.Since(t.startTime)

	if err != nil {
		t.items[index].Status = StatusFailed
		t.items[index].Error = err
	} else {
		t.items[index].Status = StatusSuccess
	}

	if !t.isTTY {
		// Non-TTY mode: print timestamped completion
		ts := time.Now().Format("15:04:05")
		sym := "+"
		status := "completed"
		if err != nil {
			sym = "x"
			status = "FAILED"
		}
		fmt.Printf("[%s] %s %s %s (%s)\n", ts, sym, t.items[index].Name, status, formatDuration(t.items[index].Duration))

		// Print error details if present
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
		}
	}
}

// Stop ends the progress tracking
func (t *Tracker) Stop() {
	close(t.stopChan)

	if t.isTTY {
		t.mu.Lock()
		if t.useColor {
			fmt.Print("\033[0m") // Ensure terminal state is reset
		}
		t.printFinal()
		t.mu.Unlock()
	}
}

// GetStatus returns a formatted status line for the current item (for TTY mode)
func (t *Tracker) GetStatus() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.current < 0 || t.current >= len(t.items) {
		return ""
	}

	item := t.items[t.current]
	if item.Status != StatusRunning {
		return ""
	}

	elapsed := time.Since(t.startTime)
	spinner := spinnerFrames[t.spinnerFrame%len(spinnerFrames)]

	displayName := item.Name
	if item.Info != "" {
		if t.useColor {
			// Dim the info in parentheses
			displayName = fmt.Sprintf("%s \033[2m(%s)\033[0m", item.Name, item.Info)
		} else {
			displayName = fmt.Sprintf("%s (%s)", item.Name, item.Info)
		}
	}

	if t.useColor {
		return fmt.Sprintf("\033[2K\r  %s %s \033[2m%s\033[0m", spinner, displayName, formatDuration(elapsed))
	}
	return fmt.Sprintf("\r  %s %s %s", spinner, displayName, formatDuration(elapsed))
}

func (t *Tracker) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case <-ticker.C:
			t.mu.Lock()
			if t.current >= 0 && t.items[t.current].Status == StatusRunning {
				t.spinnerFrame++
				elapsed := time.Since(t.startTime)
				spinner := spinnerFrames[t.spinnerFrame%len(spinnerFrames)]

				// Clear line and print status
				item := t.items[t.current]
				displayName := item.Name
				if item.Info != "" {
					if t.useColor {
						// Dim the info in parentheses
						displayName = fmt.Sprintf("%s \033[2m(%s)\033[0m", item.Name, item.Info)
					} else {
						displayName = fmt.Sprintf("%s (%s)", item.Name, item.Info)
					}
				}

				if t.useColor {
					fmt.Printf("\033[2K\r  %s %s \033[2m%s\033[0m", spinner, displayName, formatDuration(elapsed))
				} else {
					fmt.Printf("\r  %s %s %s", spinner, displayName, formatDuration(elapsed))
				}
			}
			t.mu.Unlock()
		}
	}
}

func (t *Tracker) printFinal() {
	// Clear current line
	fmt.Print("\033[2K\r")
}

// PrintItemStart prints the start of an item (used with TTY for the status line)
func (t *Tracker) PrintItemStart(index int, prefix string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTTY {
		// In TTY mode, we'll update in place
		fmt.Print(prefix)
	}
}

// PrintItemComplete prints the completion status of an item
func (t *Tracker) PrintItemComplete(index int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isTTY {
		return // Already printed in CompleteItem for non-TTY
	}

	item := t.items[index]

	// Clear the spinner line and move to new line
	fmt.Print("\033[2K\r")

	var sym string
	var suffix string

	switch item.Status {
	case StatusSuccess:
		if t.useColor {
			sym = "\033[32m+\033[0m" // green
		} else {
			sym = "+"
		}
		suffix = fmt.Sprintf("(%s)", formatDuration(item.Duration))
	case StatusFailed:
		if t.useColor {
			sym = "\033[31mx\033[0m" // red
		} else {
			sym = "x"
		}
		suffix = fmt.Sprintf("(%s) FAILED", formatDuration(item.Duration))
	}

	// Print the error details if present
	if item.Error != nil {
		// Print the completion line first, then the error
		displayName := item.Name
		if item.Info != "" {
			if t.useColor {
				displayName = fmt.Sprintf("%s \033[2m(%s)\033[0m", item.Name, item.Info)
			} else {
				displayName = fmt.Sprintf("%s (%s)", item.Name, item.Info)
			}
		}
		if t.useColor {
			suffix = fmt.Sprintf("\033[2m%s\033[0m", suffix)
		}
		fmt.Printf("  %s %s %s\n", sym, displayName, suffix)

		// Print error message
		errMsg := item.Error.Error()
		if t.useColor {
			fmt.Printf("\n\033[31m%s\033[0m\n", errMsg)
		} else {
			fmt.Printf("\n%s\n", errMsg)
		}
		return
	}

	displayName := item.Name
	if item.Info != "" {
		if t.useColor {
			// Dim the info in parentheses
			displayName = fmt.Sprintf("%s \033[2m(%s)\033[0m", item.Name, item.Info)
		} else {
			displayName = fmt.Sprintf("%s (%s)", item.Name, item.Info)
		}
	}

	// Dim the duration suffix as well
	if t.useColor {
		suffix = fmt.Sprintf("\033[2m%s\033[0m", suffix)
	}

	fmt.Printf("  %s %s %s\n", sym, displayName, suffix)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second

	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// Summary returns a summary string of completed tasks
func (t *Tracker) Summary() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var totalDuration time.Duration
	successCount := 0
	failCount := 0

	for _, item := range t.items {
		totalDuration += item.Duration
		switch item.Status {
		case StatusSuccess:
			successCount++
		case StatusFailed:
			failCount++
		}
	}

	var parts []string
	if successCount > 0 {
		parts = append(parts, fmt.Sprintf("%d succeeded", successCount))
	}
	if failCount > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failCount))
	}

	return fmt.Sprintf("%s in %s", strings.Join(parts, ", "), formatDuration(totalDuration))
}
