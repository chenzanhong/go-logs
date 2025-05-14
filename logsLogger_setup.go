package logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func (l *LogsLogger) initLoggers(output io.Writer) {
	flags := l.logFlags
	if flags&Lrootfile != 0 {
		l.hasRootFilePrefix = true

		flags = flags &^ Lrootfile // 移除 Lrootfile 标志

		// 检查并移除 Lshortfile 和 Llongfile，避免重复输出
		if flags&Lshortfile != 0 {
			flags = flags &^ Lshortfile
		}
		if flags&Llongfile != 0 {
			flags = flags &^ Llongfile
		}
	}

	var multiWriter io.Writer
	if output != os.Stderr {
		multiWriter = io.MultiWriter(os.Stderr, output)
	}

	// 初始化每个级别的日志器
	l.debugL = log.New(output, "[DEBUG] ", flags)
	l.infoL = log.New(output, "[INFO] ", flags)
	l.warnL = log.New(output, "[WARN] ", flags)
	l.errorL = log.New(output, "[ERROR] ", flags)
	l.fatalL = log.New(multiWriter, "[FATAL] ", flags)
	l.panicL = log.New(multiWriter, "[PANIC] ", flags)
}

func (l *LogsLogger) initFileLog(logFilePath string) {

	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = l.logConf.MaxSize
	fileLogger.MaxBackups = l.logConf.MaxBackups
	fileLogger.MaxAge = l.logConf.KeepDays
	fileLogger.Compress = l.logConf.Compress

	l.output = fileLogger
	// 重新初始化所有日志器
	l.initLoggers(fileLogger)
}

func (l *LogsLogger) initMultiWriter(logFilePath string) {

	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = l.logConf.MaxSize
	fileLogger.MaxBackups = l.logConf.MaxBackups
	fileLogger.MaxAge = l.logConf.KeepDays
	fileLogger.Compress = l.logConf.Compress

	// 创建一个同时写入控制台和文件的 Writer
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	l.output = multiWriter
	// 重新初始化所有日志器
	l.initLoggers(multiWriter)
}

func (l *LogsLogger) SetUp(logConf LogConf) error {
	mu2.Lock()
	defer mu2.Unlock()

	// 检查日志配置是否有效
	if logConf.Mode == "" {
		logConf.Mode = defaultLogConf.Mode
	}
	if logConf.Level == 0 {
		logConf.Level = defaultLogConf.Level
	}
	if logConf.Encoding == "" {
		logConf.Encoding = defaultLogConf.Encoding
	}
	if logConf.MaxSize == 0 {
		logConf.MaxSize = defaultLogConf.MaxSize
	}
	if logConf.MaxBackups == 0 {
		logConf.MaxBackups = defaultLogConf.MaxBackups
	}
	if logConf.KeepDays == 0 {
		logConf.KeepDays = defaultLogConf.KeepDays
	}
	if logConf.Path == "" {
		logConf.Path = defaultLogConf.Path
	}

	l.logConf = logConf
	l.logFlags = LogFlagsCommon
	l.hasRootFilePrefix = false
	l.logWriteStrategy = LoggingSync // 默认同步模式

	if l.logConf.Mode == "file" || l.logConf.Mode == "both" {
		if l.logConf.Path == "" {
			return errors.New("log path is required")
		}
	}

	// 设置编码
	switch logConf.Encoding {
	case LogEncodingPlain:
		l.encoder = &PlainEncoder{}
	case LogEncodingJSON:
		l.encoder = &JsonEncoder{}
	default:
		return fmt.Errorf("unsupported log encoding: %s", logConf.Encoding)
	}

	// 设置日志级别
	if LogLevel(logConf.Level) < LogLevelDebug {
		return errors.New("invalid log level")
	}

	// currentLogLevel = LogLevel(logConf.Level)

	// 获取项目根目录
	projectRootOnce.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	// 初始化输出
	switch l.logConf.Mode {
	case "file":
		l.initFileLog(l.logConf.Path)
	case "both":
		l.initMultiWriter(l.logConf.Path)
	default:
		l.initLoggers(os.Stdout)
	}

	return nil
}

// SetOutput 设置日志输出位置，自动更新Mode
func (l *LogsLogger) SetOutput(writer io.Writer) error {
	mu2.Lock()
	defer mu2.Unlock()

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	l.output = writer

	mode := LogModeConsole

	switch w := writer.(type) {
	case *os.File:
		// 文件输出
		if w.Name() == os.DevNull {
			mode = LogModeConsole // 特殊情况： /dev/null，仍视为console
		}else if isStdStream(w) {
			mode = LogModeConsole
		} else {
			mode = LogModeFile
			l.logConf.Path = w.Name()
		}
	case interface{ Writers() []io.Writer }:
		hasFile := false
		hasConsole := false

		for _, wr := range w.Writers() {
			if f, ok := wr.(*os.File); ok && !isStdStream(f) {
				hasFile = true
				l.logConf.Path = f.Name()
			} else if isStdStream(wr) {
				hasConsole = true
			}
		}

		if hasFile && hasConsole {
			mode = LogModeBoth
		} else if hasFile {
			mode = LogModeFile
		} else {
			mode = LogModeConsole
		}
	default:
		mode = LogModeConsole
	}

	fmt.Println("mode：", mode)

	l.logConf.Mode = mode
	initLoggers(l.output)

	return nil
}

// 设置编码
func (l *LogsLogger) SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	mu2.Lock()
	defer mu2.Unlock()
	l.logConf.Encoding = encoding

	switch encoding {
	case LogEncodingPlain:
		l.encoder = &PlainEncoder{}
	case LogEncodingJSON:
		l.encoder = &JsonEncoder{}
	default:
		return fmt.Errorf("unsupported log encoding: %s", encoding)
	}
	return nil
}

// 设置日志文件最大大小
func (l *LogsLogger) SetMaxSize(maxSize int) {
	mu2.Lock()
	defer mu2.Unlock()
	l.logConf.MaxSize = maxSize

	// 重新初始化日志器以应用新设置
	if l.logConf.Mode == "file" {
		l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		l.initMultiWriter(l.logConf.Path)
	}
}

// 设置日志文件最大保留天数
func (l *LogsLogger) SetMaxAge(maxAge int) {
	mu2.Lock()
	defer mu2.Unlock()
	l.logConf.KeepDays = maxAge

	// 重新初始化日志器以应用新设置
	if l.logConf.Mode == "file" {
		l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		l.initMultiWriter(l.logConf.Path)
	}
}

// 设置日志文件最大保留数量
func (l *LogsLogger) SetMaxBackups(maxBackups int) {
	mu2.Lock()
	defer mu2.Unlock()
	l.logConf.MaxBackups = maxBackups

	if l.logConf.Mode == "file" {
		l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		l.initMultiWriter(l.logConf.Path)
	}
}

func (l *LogsLogger) SetLogLevel(level LogLevel) error {
	mu2.Lock()
	defer mu2.Unlock()

	if level < LogLevelDebug {
		return errors.New("invalid log level")
	}

	l.logConf.Level = int(level)
	return nil
}

// 设置标志
func (l *LogsLogger) SetFlags(flags int) error {
	mu2.Lock()
	defer mu2.Unlock()

	// 对flags的合法性进行检查
	// 检查是否设置了无效的标志
	const vaildFlags = Ldate | Ltime | Lmicroseconds | Llongfile | Lshortfile | LUTC | Lmsgprefix | Lrootfile
	if flags < 0 || (flags & ^vaildFlags) != 0 {
		return errors.New("invalid flags value")
	}

	// 检查是否设置了 Ldate、Ltime 或 Lmicroseconds 标志
	if flags&(Ldate|Ltime|Lmicroseconds) == 0 {
		// 如果没有设置日期、时间或微秒，设置默认的 Ldate | Ltime
		flags = Ldate | Ltime
	}

	// 检查是否设置了 Lrootfile 标志
	if flags&Lrootfile != 0 {
		l.hasRootFilePrefix = true
		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
		
		// 检查并移除 Lshortfile 和 Llongfile，避免重复输出
		if flags&Lshortfile != 0 {
			flags = flags &^ Lshortfile
		}
		if flags&Llongfile != 0 {
			flags = flags &^ Llongfile
		}
	}

	l.logFlags = flags

	l.debugL.SetFlags(flags)
	l.infoL.SetFlags(flags)
	l.warnL.SetFlags(flags)
	l.errorL.SetFlags(flags)
	l.fatalL.SetFlags(flags)
	l.panicL.SetFlags(flags)
	return nil
}

// 设置日志同步还是异步
func (l *LogsLogger) SetLogWriteStrategy(strategy logWriteStrategy) {
	mu2.Lock()
	defer mu2.Unlock()
	l.logWriteStrategy = strategy
}

// 设置前缀
func (l *LogsLogger) SetPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.debugL.SetPrefix("[DEBUG] " + prefix)
	l.infoL.SetPrefix("[INFO] " + prefix)
	l.warnL.SetPrefix("[WARN] " + prefix)
	l.errorL.SetPrefix("[ERROR] " + prefix)
	l.fatalL.SetPrefix("[FATAL] " + prefix)
	l.panicL.SetPrefix("[PANIC] " + prefix)
}

func (l *LogsLogger) SetDebugPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.debugL.SetPrefix(prefix)
}

func (l *LogsLogger) SetDebugPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.debugL.SetPrefix("[DEBUG] " + prefix)
}

func (l *LogsLogger) SetInfoPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.infoL.SetPrefix(prefix)
}

func (l *LogsLogger) SetInfoPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.infoL.SetPrefix("[INFO] " + prefix)
}

func (l *LogsLogger) SetWarnPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.warnL.SetPrefix(prefix)
}

func (l *LogsLogger) SetWarnPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.warnL.SetPrefix("[WARN] " + prefix)
}

func (l *LogsLogger) SetErrorPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.errorL.SetPrefix(prefix)
}

func (l *LogsLogger) SetErrorPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.errorL.SetPrefix("[ERROR] " + prefix)
}

func (l *LogsLogger) SetFatalPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.fatalL.SetPrefix(prefix)
}

func (l *LogsLogger) SetFatalPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.fatalL.SetPrefix("[FATAL] " + prefix)
}

func (l *LogsLogger) SetPanicPrefixWithoutDefaultPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.panicL.SetPrefix(prefix)
}

func (l *LogsLogger) SetPanicPrefix(prefix string) {
	mu2.Lock()
	defer mu2.Unlock()
	l.panicL.SetPrefix("[PANIC] " + prefix)
}
