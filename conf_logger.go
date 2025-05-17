package logs

import (
	"bytes"
	"os"
	"sync"
	"time"
)

func NewDefaultLogConf() LogConf {
	return LogConf{ // 默认日志配置
		Mode:       "console",         // 默认输出到控制台,
		Level:      int(LogLevelInfo), // 默认日志级别为 INFO
		Encoding:   "plain",           // 默认编码为 plain text
		Path:       "",                // 控制台模式下不需要路径
		MaxSize:    1,                 // 默认每个日志文件最大 10MB
		MaxBackups: 3,                 // 默认最多保留 3 个备份
		KeepDays:   1,                 // 默认日志文件保留 30 天
		Compress:   false,             // 默认不压缩旧的日志文件
	}
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
	if custom.Level >= 0 && custom.Level <= 5 {
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
		logWriteStrategy:  LoggingSync,
		logChan:           make(chan *logItem, defaultLogChanSize),
		shutdownChan:      make(chan struct{}),
		itemPool: sync.Pool{
			New: func() interface{} {
				return &logItem{}
			},
		},
		batchBuffer: make([][]byte, 0, batchSize),
		batchTicker: time.NewTicker(flushInterval),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}

	// 获取项目根目录
	projectRootOnce.Do(func() {
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
	err := logger.Setup(conf)
	if err != nil {
		return nil, err
	}
	return logger, nil
}
