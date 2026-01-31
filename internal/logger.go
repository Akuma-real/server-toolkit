package internal

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger 日志记录器
type Logger struct {
	level  LogLevel
	output *os.File
	mu     sync.Mutex
}

// NewLogger 创建新日志记录器
func NewLogger(level LogLevel, output *os.File) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	prefix := fmt.Sprintf("[%s] [%s] ",
		time.Now().Format("2006-01-02 15:04:05"),
		levelName(level),
	)
	message := fmt.Sprintf(format, args...)

	fmt.Fprintf(l.output, "%s%s\n", prefix, message)
}

// Debug 记录 DEBUG 级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 记录 INFO 级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 记录 WARN 级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 记录 ERROR 级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// levelName 返回日志级别名称
func levelName(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}
