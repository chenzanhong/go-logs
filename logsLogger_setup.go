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
	}
	// 初始化每个级别的日志器
	l.debugL = log.New(output, "[DEBUG] ", flags)
	l.infoL = log.New(output, "[INFO] ", flags)
	l.warnL = log.New(output, "[WARN] ", flags)
	l.errorL = log.New(output, "[ERROR] ", flags)
	l.fatalL = log.New(os.Stderr, "[FATAL] ", flags)
	l.panicL = log.New(os.Stderr, "[PANIC] ", flags)
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
	mu.Lock()
	defer mu.Unlock()

	l.logConf = logConf
	l.logFlags = LogFlagsCommon
	l.hasRootFilePrefix = false

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

	currentLogLevel = LogLevel(logConf.Level)

	// 获取项目根目录
	once.Do(func() {
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

// SetOutput 设置日志输出位置
func (l *LogsLogger) SetOutput(writer io.Writer) error {
	mu.Lock()
	defer mu.Unlock()

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	switch l.logConf.Mode {
	case "file":
		if fWriter, ok := writer.(*os.File); ok {
			l.initFileLog(fWriter.Name())
		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
			for _, w := range mw.Writers() {
				if fWriter, ok := w.(*os.File); ok {
					l.initFileLog(fWriter.Name())
					break
				}
			}
		} else {
			return errors.New("unsupported writer type for file mode")
		}
	case "both":
		// 如果是 both 模式，调用 l.initMultiWriter 并传递 writer 中的路径
		if fWriter, ok := writer.(*os.File); ok {
			l.initMultiWriter(fWriter.Name())
		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
			// 如果是 MultiWriter，尝试从中找到 *os.File
			for _, w := range mw.Writers() {
				if fWriter, ok := w.(*os.File); ok {
					l.initMultiWriter(fWriter.Name())
					break
				}
			}
		} else {
			return errors.New("unsupported writer type for both mode")
		}
	case "console":
		// 如果是 console 模式，直接设置 l.output
		l.output = writer
		l.initLoggers(l.output)
	default:
		return errors.New("unsupported log mode")
	}

	return nil
}

// 设置编码
func (l *LogsLogger) SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
	l.logConf.MaxBackups = maxBackups

	if l.logConf.Mode == "file" {
		l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		l.initMultiWriter(l.logConf.Path)
	}
}

func (l *LogsLogger) SetLogLevel(level LogLevel) error {
	mu.Lock()
	defer mu.Unlock()

	if level < LogLevelDebug {
		return errors.New("invalid log level")
	}

	currentLogLevel = level
	return nil
}

// 设置标志
func (l *LogsLogger) SetFlags(flags int) error {
	mu.Lock()
	defer mu.Unlock()

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

// 设置前缀
func (l *LogsLogger) SetPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	l.debugL.SetPrefix(prefix)
	l.infoL.SetPrefix(prefix)
	l.warnL.SetPrefix(prefix)
	l.errorL.SetPrefix(prefix)
	l.fatalL.SetPrefix(prefix)
	l.panicL.SetPrefix(prefix)
}
func (l *LogsLogger) SetDebugPrefix(prefix string) {
	l.debugL.SetPrefix(prefix)
}
func (l *LogsLogger) SetInfoPrefix(prefix string) {
	l.infoL.SetPrefix(prefix)
}
func (l *LogsLogger) SetWarnPrefix(prefix string) {
	l.warnL.SetPrefix(prefix)
}
func (l *LogsLogger) SetErrorPrefix(prefix string) {
	l.errorL.SetPrefix(prefix)
}
func (l *LogsLogger) SetFatalPrefix(prefix string) {
	l.fatalL.SetPrefix(prefix)
}
func (l *LogsLogger) SetPanicPrefix(prefix string) {
	l.panicL.SetPrefix(prefix)
}
