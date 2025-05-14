package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// findProjectRoot 查找项目的根目录（假设存在 go.mod 文件）
func findProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("无法获取当前文件信息")
	}
	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir { // 到达根目录
			break
		}
		dir = parentDir
	}

	return "", fmt.Errorf("未能找到项目根目录（go.mod 文件）")
}

// GetRelativePath 获取调用者的相对路径和行号
func GetRelativePath(skip int) (file string, line int) {
	projectRootOnce.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	_, path, line, _ := runtime.Caller(skip)
	relativePath, err := filepath.Rel(projectRoot, path)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return path, line
	}

	return relativePath, line
}

func GetLogPrefix(skip int) (logPrefix string) {
	_, path, line, _ := runtime.Caller(skip)
	relativePath, err := filepath.Rel(projectRoot, path)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return fmt.Sprintf("%s %d: ", path, line)
	}

	return fmt.Sprintf("%s %d: ", relativePath, line)
}

// 根据日志级别获取对应的log.Logger实例
func getLoggerByLevel(logger *LogsLogger, level LogLevel) *log.Logger {
	switch level {
	case LogLevelDebug:
		return logger.debugL
	case LogLevelInfo:
		return logger.infoL
	case LogLevelWarn:
		return logger.warnL
	case LogLevelError:
		return logger.errorL
	case LogLevelFatal:
		return logger.fatalL
	case LogLevelPanic:
		return logger.panicL
	default:
		return nil
	}
}

func containsFormatSpecifier(s string) bool {
	return regexp.MustCompile(`%(?:\.\*|\*[0-9]*|[0-9.]*[a-zA-Z])`).MatchString(s)
}

func outputLog(logger *LogsLogger, level LogLevel, skip int, format string, v ...interface{}) {
	if level < LogLevel(logger.logConf.Level) {
		return
	}

	var msg string
	if format == "" {
		msg = logger.encoder.Encode(v...)
	} else {
		msg = logger.encoder.Encode(fmt.Sprintf(format, v...))
	}
	// else if containsFormatSpecifier(format) {
	// // 如果包含格式化符号（如 %s、%d），则使用 fmt.Sprintf
	// msg = logger.encoder.Encode(fmt.Sprintf(format, v...))
	// }

	if logger.hasRootFilePrefix {
		msg = GetLogPrefix(skip) + msg
	}

	if logger.logWriteStrategy == LoggingSync || logger.logConf.Mode == LogModeConsole {
		fmt.Println("刘伟开始叫了哦，快跑啊！")
		internalLogger := getLoggerByLevel(logger, level)
		internalLogger.Output(skip, msg)
	} else {
		fmt.Println("异步输出")
		select {
		case logChan <- logItem{logger: logger, level: level, msg: msg, skip: skip + 1}:
		default:
			log.Printf("日志通道已满，暂时无法异步写入日志: %s", msg)
		}
	}
}

// output 方法的实现
// Debug 输出 DEBUG 日志
func Debug(v ...interface{}) {
	outputLog(globalLogger, LogLevelDebug, 3, "", v...)
}

func Debugf(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelDebug, 3, format, v...)
}

// Info 输出 INFO 日志
func Info(v ...interface{}) {
	outputLog(globalLogger, LogLevelInfo, 3, "", v...)
}

func Infof(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelInfo, 3, format, v...)
}

// Warn 输出 WARN 日志
func Warn(v ...interface{}) {
	outputLog(globalLogger, LogLevelInfo, 3, "", v...)
}

func Warnf(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelInfo, 3, format, v...)
}

// Error 输出 ERROR 日志
func Error(v ...interface{}) {
	outputLog(globalLogger, LogLevelError, 3, "", v...)
}

func Errorf(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelError, 3, format, v...)
}

// Fatal 输出 FATAL 日志并退出程序
func Fatal(v ...interface{}) {
	outputLog(globalLogger, LogLevelFatal, 3, "", v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelFatal, 3, format, v...)
	os.Exit(1)
}

// Panic 输出 PANIC 日志并触发 panic
func Panic(v ...interface{}) {
	outputLog(globalLogger, LogLevelPanic, 3, "", v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	outputLog(globalLogger, LogLevelPanic, 3, format, v...)
	panic(fmt.Sprintf(format, v...))
}
