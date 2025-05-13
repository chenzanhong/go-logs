package logs

import (
	"fmt"
	"os"
)

// LogsLogger 的 output 方法
func (l *LogsLogger) Debug(v ...interface{}) {
	outputLog(l, LogLevelDebug, 3, "", v...)
}
func (l *LogsLogger) Debugf(format string, v ...interface{}) {
	outputLog(l, LogLevelDebug, 3, format, v...)
}

func (l *LogsLogger) Info(v ...interface{}) {
	outputLog(l, LogLevelInfo, 3, "", v...)
}
func (l *LogsLogger) Infof(format string, v ...interface{}) {
	outputLog(l, LogLevelInfo, 3, format, v...)
}

func (l *LogsLogger) Warn(v ...interface{}) {
	outputLog(l, LogLevelWarn, 3, "", v...)
}
func (l *LogsLogger) Warnf(format string, v ...interface{}) {
	outputLog(l, LogLevelWarn, 3, format, v...)
}

func (l *LogsLogger) Error(v ...interface{}) {
	outputLog(l, LogLevelError, 3, "", v...)
}
func (l *LogsLogger) Errorf(format string, v ...interface{}) {
	outputLog(l, LogLevelError, 3, format, v...)
}

func (l *LogsLogger) Fatal(v ...interface{}) {
	outputLog(l, LogLevelFatal, 3, "", v...)
	os.Exit(1)
}
func (l *LogsLogger) Fatalf(format string, v ...interface{}) {
	outputLog(l, LogLevelFatal, 3, format, v...)
	os.Exit(1)
}

func (l *LogsLogger) Panic(v ...interface{}) {
	outputLog(l, LogLevelPanic, 3, "", v...)
	panic(fmt.Sprint(v...))
}

func (l *LogsLogger) Panicf(format string, v ...interface{}) {
	outputLog(l, LogLevelPanic, 3, format, v...)
	panic(fmt.Sprintf(format, v...))
}
