package systemd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"sdtop/internal/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

// LogMsg is a Bubble Tea message containing a log entry
type LogMsg types.LogEntry

// ErrorMsg is a Bubble Tea message containing an error
type ErrorMsg string

// LogReader streams journald logs for a specific service
type LogReader struct {
	journal *sdjournal.Journal
}

// NewLogReader creates a new log reader
func NewLogReader() (*LogReader, error) {
	j, err := sdjournal.NewJournal()
	if err != nil {
		return nil, err
	}
	return &LogReader{journal: j}, nil
}

// Close closes the journal
func (lr *LogReader) Close() {
	if lr.journal != nil {
		lr.journal.Close()
	}
}

// StreamLogs streams logs for a service and sends them as Bubble Tea messages
func (lr *LogReader) StreamLogs(ctx context.Context, serviceName string) tea.Cmd {
	return func() tea.Msg {
		// Add match for the specific service
		if err := lr.journal.AddMatch("_SYSTEMD_UNIT=" + serviceName); err != nil {
			return ErrorMsg(fmt.Sprintf("Failed to add match: %v", err))
		}

		// Seek to end and back one entry to start from recent logs
		if err := lr.journal.SeekTail(); err != nil {
			return ErrorMsg(fmt.Sprintf("Failed to seek tail: %v", err))
		}

		// Skip back a few entries to show recent history
		for i := 0; i < 100; i++ {
			if n, err := lr.journal.Previous(); err != nil || n == 0 {
				break
			}
		}

		// Start the streaming loop
		go lr.followLogs(ctx)

		return nil
	}
}

// followLogs follows journal logs in a goroutine
func (lr *LogReader) followLogs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read next entry
			n, err := lr.journal.Next()
			if err != nil {
				continue
			}

			if n > 0 {
				entry, err := lr.journal.GetEntry()
				if err != nil {
					continue
				}

				// Extract fields
				message := entry.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE]
				priority := entry.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY]
				timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*1000)

				logEntry := types.LogEntry{
					Timestamp: timestamp,
					Message:   message,
					Priority:  priority,
				}

				// This is a simplified version - in real implementation,
				// we'd send this through a channel to the UI
				_ = logEntry
			}

			// Wait for new entries
			lr.journal.Wait(time.Millisecond * 100)
		}
	}
}

// GetRecentLogs retrieves recent logs for a service
func (lr *LogReader) GetRecentLogs(serviceName string, count int) ([]types.LogEntry, error) {
	// Clear any previous matches
	lr.journal.FlushMatches()

	// Add match for the specific service
	if err := lr.journal.AddMatch("_SYSTEMD_UNIT=" + serviceName); err != nil {
		return nil, fmt.Errorf("failed to add match: %w", err)
	}

	// Seek to tail
	if err := lr.journal.SeekTail(); err != nil {
		return nil, fmt.Errorf("failed to seek tail: %w", err)
	}

	// Go back 'count' entries
	for i := 0; i < count; i++ {
		if n, err := lr.journal.Previous(); err != nil || n == 0 {
			break
		}
	}

	var logs []types.LogEntry

	// Read forward up to 'count' entries
	for i := 0; i < count; i++ {
		n, err := lr.journal.Next()
		if err != nil || n == 0 {
			break
		}

		entry, err := lr.journal.GetEntry()
		if err != nil {
			continue
		}

		message := entry.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE]
		priorityStr := entry.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY]
		timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*1000)

		// Parse priority
		priority := "info"
		if p, err := strconv.Atoi(priorityStr); err == nil {
			switch {
			case p <= 3:
				priority = "error"
			case p <= 4:
				priority = "warn"
			default:
				priority = "info"
			}
		}

		logs = append(logs, types.LogEntry{
			Timestamp: timestamp,
			Message:   message,
			Priority:  priority,
		})
	}

	return logs, nil
}
