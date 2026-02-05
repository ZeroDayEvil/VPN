package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	maxLogSize  = 5 * 1024 * 1024 // 5 MB
	maxLogFiles = 5
)

type Logger struct {
	logger   *log.Logger
	file     *os.File
	logPath  string
	fileSize int64
}

func NewLogger(logPath string) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "", log.LstdFlags)

	l := &Logger{
		logger:   logger,
		file:     file,
		logPath:  logPath,
		fileSize: stat.Size(),
	}

	return l, nil
}

func (l *Logger) checkRotation() {
	if l.fileSize >= maxLogSize {
		l.rotate()
	}
}

func (l *Logger) rotate() {
	l.file.Close()

	// Rotate log files
	for i := maxLogFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)
		os.Rename(oldPath, newPath)
	}

	// Move current log to .1
	os.Rename(l.logPath, l.logPath+".1")

	// Create new log file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	l.logger = log.New(multiWriter, "", log.LstdFlags)
	l.file = file
	l.fileSize = 0
}

func (l *Logger) write(level string, format string, v ...interface{}) {
	l.checkRotation()
	msg := fmt.Sprintf(format, v...)
	logMsg := fmt.Sprintf("[%s] %s", level, msg)
	l.logger.Println(logMsg)
	l.fileSize += int64(len(logMsg) + 1)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.write("INFO", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.write("ERROR", format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
	l.write("WARN", format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.write("DEBUG", format, v...)
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
