package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	once.Do(func() {
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

func outputLog(logger *log.Logger, skip int, v ...interface{}) {
	if currentLogLevel > LogLevel(logConfig.Level) {
		return
	}
	msg := encoder.Encode(v...)
	if rootFilePrefix {
		msg = GetLogPrefix(skip) + msg
	}
	logger.Output(skip, msg)
}

func outputLogf(logger *log.Logger, skip int, format string, v ...interface{}) {
	if currentLogLevel > LogLevel(logConfig.Level) {
		return
	}
	msg := encoder.Encode(fmt.Sprintf(format, v...))
	if rootFilePrefix {
		msg = GetLogPrefix(skip) + msg
	}
	logger.Output(skip, msg)
}

// Info 输出 INFO 日志
func Info(v ...interface{}) {
	outputLog(infoLogger, 3, v...)
}

func Infof(format string, v ...interface{}) {
	outputLogf(infoLogger, 3, format, v...)
}

// Warn 输出 WARN 日志
func Warn(v ...interface{}) {
	outputLog(warnLogger, 3, v...)
}

func Warnf(format string, v ...interface{}) {
	outputLogf(warnLogger, 3, format, v...)
}

// Error 输出 ERROR 日志
func Error(v ...interface{}) {
	outputLog(errorLogger, 3, v...)
}

func Errorf(format string, v ...interface{}) {
	outputLogf(errorLogger, 3, format, v...)
}

// Fatal 输出 FATAL 日志并退出程序
func Fatal(v ...interface{}) {
	outputLog(fatalLogger, 3, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	outputLogf(fatalLogger, 3, format, v...)
	os.Exit(1)
}

// Panic 输出 PANIC 日志并触发 panic
func Panic(v ...interface{}) {
	outputLog(panicLogger, 3, v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	outputLogf(panicLogger, 3, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	outputLog(debugLogger, 3, v...)
}

func Debugf(format string, v ...interface{}) {
	outputLogf(debugLogger, 3, format, v...)
}
