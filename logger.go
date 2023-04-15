package main

import (
	"io"
	"log"
	"os"
)

type LogLevel int

var (
	GlobalLogger *Logger
)

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
)

type Logger struct {
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
	debugLog     *log.Logger
	infoLog      *log.Logger
	warningLog   *log.Logger
	errorLog     *log.Logger
}

func NewLogger(debugHandle, infoHandle, warningHandle, errorHandle io.Writer) *Logger {
	return &Logger{
		debugEnabled: os.Getenv("MW_AGENT_DEBUG_LOGS") != "",
		infoEnabled:  os.Getenv("MW_AGENT_INFO_LOGS") != "",
		warnEnabled:  os.Getenv("MW_AGENT_WARN_LOGS") != "",
		errorEnabled: os.Getenv("MW_AGENT_ERROR_LOGS") != "",
		debugLog:     log.New(debugHandle, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:      log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLog:   log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog:     log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) SetOutput(debugHandle, infoHandle, warningHandle, errorHandle io.Writer) {
	l.debugLog.SetOutput(debugHandle)
	l.infoLog.SetOutput(infoHandle)
	l.warningLog.SetOutput(warningHandle)
	l.errorLog.SetOutput(errorHandle)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.debugEnabled {
		l.debugLog.Println(v...)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.infoEnabled {
		l.infoLog.Println(v...)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.warnEnabled {
		l.warningLog.Println(v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.errorEnabled {
		l.errorLog.Println(v...)
	}
}

func initLogger() {
	debugLogHandle := os.Stdout
	infoLogHandle := os.Stdout
	warnLogHandle := os.Stdout
	errorLogHandle := os.Stderr

	GlobalLogger = NewLogger(debugLogHandle, infoLogHandle, warnLogHandle, errorLogHandle)
}
