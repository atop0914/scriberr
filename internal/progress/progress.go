package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar implements a simple progress bar
type ProgressBar struct {
	total     int
	current   int
	message   string
	width     int
	mu        sync.Mutex
	startTime time.Time
	lastUpdate time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:     total,
		current:   0,
		message:   message,
		width:     40,
		startTime: time.Now(),
	}
}

// Start begins the progress tracking
func (p *ProgressBar) Start(total int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
	p.current = 0
	p.message = message
	p.startTime = time.Now()
	p.render()
}

// Update updates the current progress
func (p *ProgressBar) Update(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = current
	p.lastUpdate = time.Now()
	p.render()
}

// Complete marks the progress as complete
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = p.total
	p.render()
	p.println("")
}

// Error reports an error during progress
func (p *ProgressBar) Error(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.println("")
	fmt.Printf("  \033[31m✗ Error: %v\033[0m\n", err)
}

// render draws the progress bar to stdout
func (p *ProgressBar) render() {
	if p.total == 0 {
		return
	}

	// Calculate percentage and filled width
	percent := float64(p.current) / float64(p.total)
	filled := int(float64(p.width) * percent)

	// Build progress bar string
	bar := strings.Repeat("=", filled)
	if filled < p.width {
		bar += ">"
		bar += strings.Repeat(" ", p.width-filled-1)
	}

	// Calculate elapsed time
	elapsed := time.Since(p.startTime)

	// Estimate remaining time
	var remaining time.Duration
	if p.current > 0 {
		avgTimePerUnit := elapsed / time.Duration(p.current)
		remaining = avgTimePerUnit * time.Duration(p.total-p.current)
	}

	// Move cursor to beginning of line and clear
	fmt.Printf("\r\033[2K")

	// Print progress bar
	fmt.Printf("  \033[36m%s\033[0m ", p.message)
	fmt.Printf("[%s] \033[32m%d%%\033[0m", bar, int(percent*100))

	if remaining > 0 {
		fmt.Printf(" ETA: %v", remaining.Truncate(time.Second))
	} else if p.current == p.total {
		fmt.Printf(" %v", elapsed.Truncate(time.Second))
	}
}

// Println prints a line (helper for error messages)
func (p *ProgressBar) Println(args ...interface{}) {
	p.println(fmt.Sprint(args...))
}

func (p *ProgressBar) println(s string) {
	fmt.Printf("\r\033[2K%s\n", s)
}

// SilentProgress is a no-op progress tracker
type SilentProgress struct{}

// Start does nothing
func (s *SilentProgress) Start(total int, message string) {}

// Update does nothing
func (s *SilentProgress) Update(current int) {}

// Complete does nothing
func (s *SilentProgress) Complete() {}

// Error does nothing
func (s *SilentProgress) Error(err error) {}
