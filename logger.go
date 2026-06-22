package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var once sync.Once
var currentFile *os.File
var mu sync.Mutex

func Init(logDir string, env string) {
	once.Do(func() {

		// Create log directory
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create log directory: %v", err))
		}

		updateLogTarget(logDir, env)

		// Wait to next midnight
		go func() {
			for {
				now := time.Now()
				nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

				time.Sleep(time.Until(nextMidnight))

				updateLogTarget(logDir, env)
			}
		}()
	})
}

func updateLogTarget(logDir string, env string) {
	mu.Lock()
	defer mu.Unlock()

	// Path as /log/directory/ENV/YYYY-MM-DD.log
	logPath := filepath.Join(logDir, env, time.Now().Format("2006-01-02")+".log")

	// Ensure environment subdirectory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		fmt.Printf("Failed to create env log directory %s: %v", filepath.Dir(logPath), err)
		return
	}

	newFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file %s: %v", logPath, err)
		return
	}

	logOutput := io.MultiWriter(os.Stdout, newFile)

	handler := slog.NewJSONHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	slog.SetDefault(slog.New(handler))

	if currentFile != nil {
		currentFile.Close()
	}
	currentFile = newFile
}
