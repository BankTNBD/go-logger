package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// LogMessage represents the expected JSON structure of our log entries
type LogMessage struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Env   string `json:"env"`
}

func TestLoggerInitialization(t *testing.T) {
	// 1. Create a memory buffer to intercept log outputs
	var buf bytes.Buffer

	// 2. Explicitly initialize a JSON handler pointing to our buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	// 3. Write a sample log entry
	expectedMsg := "testing reusable logger framework"
	slog.Info(expectedMsg, "env", "test-env")

	// 4. Parse and validate the output buffer
	var logEntry LogMessage
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	// 5. Assertions
	if logEntry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", logEntry.Level)
	}

	if logEntry.Msg != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, logEntry.Msg)
	}

	if logEntry.Env != "test-env" {
		t.Errorf("Expected context field env to be 'test-env', got %q", logEntry.Env)
	}
}

func TestUpdateLogTargetCreatesLogFileAndWritesEntry(t *testing.T) {
	// Create temp dir
	td := t.TempDir()
	env := "testenv"

	// Call updateLogTarget to set up logging to the file
	updateLogTarget(td, env)

	// Write a log entry; the handler set by updateLogTarget writes to the file
	expected := "entry-from-test"
	slog.Info(expected, "env", env)

	// Give the logger a moment to flush to disk
	time.Sleep(10 * time.Millisecond)

	// Read the log file
	logPath := filepath.Join(td, env, time.Now().Format("2006-01-02")+".log")
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file %s: %v", logPath, err)
	}

	// The file may contain multiple JSON objects; look for our message
	if !bytes.Contains(b, []byte(expected)) {
		t.Fatalf("log file did not contain expected message %q; contents=%s", expected, string(b))
	}
}
