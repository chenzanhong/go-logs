package logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logConfig LogConf // 日志配置

	// 日志器
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	panicLogger *log.Logger

	fileLogger *lumberjack.Logger // 用于文件输出的日志器

	encoder Encoder = &PlainEncoder{} // 默认使用 PlainEncoder

	logFlags = LogFlagsCommon // 默认日志标志

	currentLogLevel LogLevel = LogLevelInfo // 默认日志级别为 Info

	logOutput io.Writer = os.Stdout // 默认输出到控制台

	rootFilePrefix bool = false // 自定义的相对路径前缀
	projectRoot    string
	mu             sync.Mutex
	once           sync.Once
)

// InitLogger 初始化日志记录器
// 可以根据需要调整日志级别和输出位置
func initLoggers(output io.Writer) {
	flags := logFlags
	if flags&Lrootfile != 0 {
		rootFilePrefix = true
		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
	}
	debugLogger = log.New(output, "[DEBUG] ", flags)
	infoLogger = log.New(output, "[INFO] ", flags)
	warnLogger = log.New(output, "[WARN] ", flags)
	errorLogger = log.New(output, "[ERROR] ", flags)
	fatalLogger = log.New(os.Stderr, "[FATAL] ", flags)
	panicLogger = log.New(os.Stderr, "[PANIC] ", flags)
}

// initFileLog 初始化日志文件输出
func initFileLog(logFilePath string) {
	mu.Lock()
	defer mu.Unlock()

	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = logConfig.MaxSize
	fileLogger.MaxBackups = logConfig.MaxBackups
	fileLogger.MaxAge = logConfig.KeepDays
	fileLogger.Compress = logConfig.Compress

	// 重新初始化所有日志器
	initLoggers(fileLogger)
}

// initMultiWriter 初始化同时输出到控制台和文件的日志器
func initMultiWriter(logFilePath string) {
	mu.Lock()
	defer mu.Unlock()

	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = logConfig.MaxSize
	fileLogger.MaxBackups = logConfig.MaxBackups
	fileLogger.MaxAge = logConfig.KeepDays
	fileLogger.Compress = logConfig.Compress

	// 创建一个同时写入控制台和文件的 Writer
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// 重新初始化所有日志器
	initLoggers(multiWriter)
}

// DefaultLogConf 返回一个带有默认配置的 LogConf 实例
func DefaultLogConf() LogConf {
	return LogConf{
		Mode:       "console",         // 默认输出到控制台
		Level:      int(LogLevelInfo), // 默认日志级别为 Info
		Encoding:   "plain",           // 默认编码为 plain text
		Path:       "",                // 控制台模式下不需要路径
		MaxSize:    10,                // 每个日志文件最大 10MB（仅当 Mode 为 file 或 both 时有效）
		MaxBackups: 3,                 // 最多保留 3 个备份（仅当 Mode 为 file 或 both 时有效）
		KeepDays:   7,                 // 日志文件保留 7 天（仅当 Mode 为 file 或 both 时有效）
		Compress:   false,             // 压缩旧的日志文件（仅当 Mode 为 file 或 both 时有效）
	}
}

// 设置方法 -----------------------------------------------------------------------
// SetUp 初始化日志记录器
func SetUp(logConf LogConf) error {
	mu.Lock()
	defer mu.Unlock()

	logConfig = logConf

	if logConfig.Mode == "file" || logConfig.Mode == "both" {
		if logConfig.Path == "" {
			return errors.New("log path is required")
		}
	}

	// 设置编码
	if err := SetEncoding(logConfig.Encoding); err != nil {
		return fmt.Errorf("failed to set encoding: %v", err)
	}

	// 设置日志级别
	if err := SetLogLevel(LogLevel(logConf.Level)); err != nil {
		return fmt.Errorf("failed to set log level: %v", err)
	}

	// 获取项目根目录
	once.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	// 初始化输出
	switch logConfig.Mode {
	case "file":
		initFileLog(logConfig.Path)
	case "both":
		initMultiWriter(logConfig.Path)
	default:
		initLoggers(os.Stdout)
	}

	return nil
}

// SetupDefault 使用默认配置初始化日志记录器
func SetupDefault() error {
	defaultConfig := DefaultLogConf()
	err := SetUp(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to set up default logger: %v", err)
	}
	return nil
}

// SetOutput 设置日志输出位置
func SetOutput(writer io.Writer) error {
	mu.Lock()
	defer mu.Unlock()

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	switch logConfig.Mode {
	case "file":
		if fWriter, ok := writer.(*os.File); ok {
			initFileLog(fWriter.Name())
		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
			for _, w := range mw.Writers() {
				if fWriter, ok := w.(*os.File); ok {
					initFileLog(fWriter.Name())
					break
				}
			}
		} else {
			return errors.New("unsupported writer type for file mode")
		}
	case "both":
		// 如果是 both 模式，调用 initMultiWriter 并传递 writer 中的路径
		if fWriter, ok := writer.(*os.File); ok {
			initMultiWriter(fWriter.Name())
		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
			// 如果是 MultiWriter，尝试从中找到 *os.File
			for _, w := range mw.Writers() {
				if fWriter, ok := w.(*os.File); ok {
					initMultiWriter(fWriter.Name())
					break
				}
			}
		} else {
			return errors.New("unsupported writer type for both mode")
		}
	case "console":
		// 如果是 console 模式，直接设置 logOutput
		logOutput = writer
		initLoggers(logOutput)
	default:
		return errors.New("unsupported log mode")
	}

	return nil
}

// 设置编码
func SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	mu.Lock()
	defer mu.Unlock()
	logConfig.Encoding = encoding

	switch encoding {
	case LogEncodingPlain:
		encoder = &PlainEncoder{}
	case LogEncodingJSON:
		encoder = &JsonEncoder{}
	default:
		return fmt.Errorf("unsupported log encoding: %s", encoding)
	}
	return nil
}

// 设置日志文件最大大小
func SetMaxSize(maxSize int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.MaxSize = maxSize

	// 重新初始化日志器以应用新设置
	if logConfig.Mode == "file" {
		initFileLog(logConfig.Path)
	} else if logConfig.Mode == "both" {
		initMultiWriter(logConfig.Path)
	}
}

// 设置日志文件最大保留天数
func SetMaxAge(maxAge int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.KeepDays = maxAge

	// 重新初始化日志器以应用新设置
	if logConfig.Mode == "file" {
		initFileLog(logConfig.Path)
	} else if logConfig.Mode == "both" {
		initMultiWriter(logConfig.Path)
	}
}

// 设置日志文件最大保留数量
func SetMaxBackups(maxBackups int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.MaxBackups = maxBackups

	if logConfig.Mode == "file" {
		initFileLog(logConfig.Path)
	} else if logConfig.Mode == "both" {
		initMultiWriter(logConfig.Path)
	}
}

func SetLogLevel(level LogLevel) error {
	mu.Lock()
	defer mu.Unlock()

	if level < LogLevelDebug {
		return errors.New("invalid log level")
	}

	currentLogLevel = level
	return nil
}

// 设置标志
func SetFlags(flags int) error {
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
		rootFilePrefix = true
		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
	}

	logFlags = flags

	debugLogger.SetFlags(flags)
	infoLogger.SetFlags(flags)
	warnLogger.SetFlags(flags)
	errorLogger.SetFlags(flags)
	fatalLogger.SetFlags(flags)
	panicLogger.SetFlags(flags)
	return nil
}

// 设置前缀
func SetPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	debugLogger.SetPrefix(prefix)
	infoLogger.SetPrefix(prefix)
	warnLogger.SetPrefix(prefix)
	errorLogger.SetPrefix(prefix)
	fatalLogger.SetPrefix(prefix)
	panicLogger.SetPrefix(prefix)
}
func SetDebugPrefix(prefix string) {
	debugLogger.SetPrefix(prefix)
}
func SetInfoPrefix(prefix string) {
	infoLogger.SetPrefix(prefix)
}
func SetWarnPrefix(prefix string) {
	warnLogger.SetPrefix(prefix)
}
func SetErrorPrefix(prefix string) {
	errorLogger.SetPrefix(prefix)
}
func SetFatalPrefix(prefix string) {
	fatalLogger.SetPrefix(prefix)
}
func SetPanicPrefix(prefix string) {
	panicLogger.SetPrefix(prefix)
}
func SetLoggerPrefix(logger *log.Logger, newPrefix string) {
	logger.SetPrefix(newPrefix)
}
