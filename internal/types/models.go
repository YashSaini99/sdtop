package types

import "time"

// Service represents a systemd service unit
type Service struct {
	Name          string
	Description   string
	ActiveState   string // active, inactive, failed
	SubState      string // running, exited, dead
	LoadState     string // loaded, not-found, masked
	UnitFileState string // enabled, disabled, static, masked
}

// LogEntry represents a single journald log entry
type LogEntry struct {
	Timestamp time.Time
	Message   string
	Priority  string
}

// Process represents a running process
type Process struct {
	PID      int
	Name     string
	Cmdline  string
	Parent   int
	Children []*Process
}
