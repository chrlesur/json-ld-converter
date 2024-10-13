package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Initialize logger
	Init(INFO, "")
	SetDebugMode(true)

	// Test logging
	Debug("This is a debug message")
	Info("This is an info message")
	Warning("This is a warning message")
	Error("This is an error message")

	// Reset stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check if all messages are present
	if !strings.Contains(output, "DEBUG") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "WARNING") {
		t.Error("Warning message not found in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Error message not found in output")
	}

	// Test silent mode
	SetSilentMode(true)
	Info("This message should not appear")

	if strings.Contains(output, "This message should not appear") {
		t.Error("Silent mode failed")
	}

	// Clean up
	Close()
}
