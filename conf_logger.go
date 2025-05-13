package logs

import (
	"errors"
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewDefaultLogConf() LogConf {
	return defaultLogConf
}

func NewLogConfWithParams(mode string, level LogLevel, encoding string, path string, maxSize int, maxBackups int, keepDays int, compress bool) LogConf {
	return LogConf{
		Mode:       mode,
		Level:      int(level),
		Encoding:   encoding,
		Path:       path,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		KeepDays:   keepDays,
		Compress:   compress,
	}
}

func NewLogConfWithDefaults(custom LogConf) LogConf {
	// 从默认配置开始
	conf := defaultLogConf

	// 如果用户提供了特定的值，则覆盖默认值
	if custom.Mode != "" {
		conf.Mode = custom.Mode
	}
	if custom.Level != 0 { // 注意：0 是 LogLevelInfo 的默认值，确保你的逻辑正确处理这种情况
		conf.Level = custom.Level
	}
	if custom.Encoding != "" {
		conf.Encoding = custom.Encoding
	}
	if custom.Path != "" {
		conf.Path = custom.Path
	}
	if custom.MaxSize != 0 {
		conf.MaxSize = custom.MaxSize
	}
	if custom.MaxBackups != 0 {
		conf.MaxBackups = custom.MaxBackups
	}
	if custom.KeepDays != 0 {
		conf.KeepDays = custom.KeepDays
	}
	if custom.Compress {
		conf.Compress = custom.Compress
	}

	return conf
}

func NewDefaultLogger() *LogsLogger {
	var logger *LogsLogger = &LogsLogger{
		encoder:           &PlainEncoder{},
		output:            os.Stdout,
		logFlags:          LogFlagsCommon,
		hasRootFilePrefix: false,
		logConf:           defaultLogConf,
	}

	// 获取项目根目录
	once.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	logger.initLoggers(os.Stdout)
	return logger
}

// NewLogger 函数用于创建一个新的日志器实例
func NewLogger(conf LogConf) (*LogsLogger, error) {
	var logger *LogsLogger = &LogsLogger{}
	err := logger.SetUp(conf)
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func newLogger(writer io.Writer, flag int, prefixFormat string) (*LogsLogger, error) {
	if writer == nil {
		return nil, errors.New("writer cannot be nil")
	}

	logger := &LogsLogger{}

	if flag&Lrootfile != 0 {
		logger.hasRootFilePrefix = true
		flag = flag &^ Lrootfile // 移除 Lrootfile 标志
	}

	// 初始化每个级别的日志器
	logger.debugL = log.New(writer, "DEBUG: ", flag)
	logger.infoL = log.New(writer, "INFO: ", flag)
	logger.warnL = log.New(writer, "WARN: ", flag)
	logger.errorL = log.New(writer, "ERROR: ", flag)
	logger.fatalL = log.New(writer, "FATAL: ", flag)
	logger.panicL = log.New(writer, "PANIC: ", flag)

	return logger, nil
}

func NewFileLogger(filename string, flag int) (*LogsLogger, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    defaultLogConf.MaxSize,
		MaxBackups: defaultLogConf.MaxBackups,
		MaxAge:     defaultLogConf.KeepDays,
		Compress:   defaultLogConf.Compress,
	}

	return newLogger(writer, flag, defaultLogConf.Encoding)
}

func NewMultiWriterLogger(filename string, flag int) (*LogsLogger, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    defaultLogConf.MaxSize,
		MaxBackups: defaultLogConf.MaxBackups,
		MaxAge:     defaultLogConf.KeepDays,
		Compress:   defaultLogConf.Compress,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

	return newLogger(multiWriter, flag, defaultLogConf.Encoding)
}
