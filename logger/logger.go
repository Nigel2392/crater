package logger

import (
	"bytes"
	"fmt"
)

type Logger interface {
	Loglevel(level LogLevel)
	Error(...any)
	Info(...any)
	Debug(...any)
}

type LogLevel string

const (
	Silent LogLevel = "silent"
	Error  LogLevel = "error"
	Info   LogLevel = "info"
	Debug  LogLevel = "debug"
)

func logLevelToInt(level LogLevel) int {
	switch level {
	case Silent:
		return 0
	case Error:
		return 1
	case Info:
		return 2
	case Debug:
		return 3
	}
	return 1
}

type SimpleLogger struct {
	ErrorFunc func(...any)
	InfoFunc  func(...any)
	DebugFunc func(...any)
	Level     int
}

func New(errorFunc, infoFunc, debugFunc func(...any)) *SimpleLogger {
	return &SimpleLogger{errorFunc, infoFunc, debugFunc, logLevelToInt("")}
}

func (l *SimpleLogger) Loglevel(level LogLevel) {
	l.Level = logLevelToInt(level)
}

func (l *SimpleLogger) Error(args ...any) {
	if l.Level >= 1 {
		l.ErrorFunc(args...)
	}
}

func (l *SimpleLogger) Info(args ...any) {
	if l.Level >= 2 {
		l.InfoFunc(args...)
	}
}

func (l *SimpleLogger) Debug(args ...any) {
	if l.Level >= 3 {
		l.DebugFunc(args...)
	}
}

var (
	StdLogger = New(
		logPrintln("ERROR"),
		logPrintln("INFO"),
		logPrintln("DEBUG"),
	)
)

func logPrintln(level string) func(...any) {
	return func(args ...any) {
		var b = &bytes.Buffer{}
		b.WriteString(level)
		b.WriteString(": ")
		fmt.Fprint(b, args...)
		println(b.String())
	}
}
