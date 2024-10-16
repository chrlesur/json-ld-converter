package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

var (
	logLevel        LogLevel
	logFile         *os.File
	console         io.Writer
	mu              sync.Mutex
	silentMode      bool
	debugMode       bool
	totalDocuments  int
	currentDocument int
	totalChunks     int
	currentChunk    int
)

func Init(level LogLevel, filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	logLevel = level

	if filePath != "" {
		var err error
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
	}

	console = os.Stdout
	return nil
}

func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = level
}

func SetSilentMode(silent bool) {
	mu.Lock()
	defer mu.Unlock()
	silentMode = silent
}

func SetDebugMode(debug bool) {
	mu.Lock()
	defer mu.Unlock()
	debugMode = debug
	if debug {
		logLevel = DEBUG
	}
}

func InitProgress(totalDocs int) {
	mu.Lock()
	defer mu.Unlock()
	totalDocuments = totalDocs
	currentDocument = 0
	totalChunks = 0
	currentChunk = 0
}

func SetTotalChunks(total int) {
	mu.Lock()
	defer mu.Unlock()
	totalChunks = total
	currentChunk = 0
}

func UpdateDocumentProgress() {
	mu.Lock()
	defer mu.Unlock()
	currentDocument++
	currentChunk = 0
	if !silentMode {
		fmt.Printf("\rDocument: %d/%d, Chunk: %d/%d", currentDocument, totalDocuments, currentChunk, totalChunks)
	}
}

func UpdateChunkProgress() {
	mu.Lock()
	defer mu.Unlock()
	currentChunk++
	if !silentMode {
		fmt.Printf("\rDocument: %d/%d, Chunk: %d/%d", currentDocument, totalDocuments, currentChunk, totalChunks)
	}
}

func log(level LogLevel, message string) {
	mu.Lock()
	defer mu.Unlock()

	if level < logLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, getLevelString(level), message)

	if logFile != nil {
		logFile.WriteString(logMessage)
	}

	if !silentMode {
		// Clear the current line
		fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
		// Print the log message
		fmt.Print(logMessage)
		// Print the progress on the next line
		fmt.Printf("Document: %d/%d, Chunk: %d/%d", currentDocument, totalDocuments, currentChunk, totalChunks)
	}
}

func getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func Debug(message string) {
	if debugMode {
		log(DEBUG, message)
	}
}

func Info(message string) {
	log(INFO, message)
}

func Warning(message string) {
	log(WARNING, message)
}

func Error(message string) {
	log(ERROR, message)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		logFile.Close()
	}
}
