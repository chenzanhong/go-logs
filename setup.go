package logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化日志记录器
// 可以根据需要调整日志级别和输出位置
func initLoggers(output io.Writer) {
	flags := globalLogger.logFlags
	if flags&Lrootfile != 0 {
		globalLogger.hasRootFilePrefix = true
		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
	}

	var multiWriter io.Writer
	if output != os.Stderr {
		multiWriter = io.MultiWriter(os.Stderr, output)
	}

	// 初始化每个级别的日志器
	globalLogger.debugL = log.New(output, "[DEBUG] ", flags)
	globalLogger.infoL = log.New(output, "[INFO] ", flags)
	globalLogger.warnL = log.New(output, "[WARN] ", flags)
	globalLogger.errorL = log.New(output, "[ERROR] ", flags)
	globalLogger.fatalL = log.New(multiWriter, "[FATAL] ", flags)
	globalLogger.panicL = log.New(multiWriter, "[PANIC] ", flags)
}

// initFileLog 初始化日志文件输出
func initFileLog(logFilePath string) {
	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = globalLogger.logConf.MaxSize
	fileLogger.MaxBackups = globalLogger.logConf.MaxBackups
	fileLogger.MaxAge = globalLogger.logConf.KeepDays
	fileLogger.Compress = globalLogger.logConf.Compress

	globalLogger.output = fileLogger
	// 重新初始化所有日志器
	initLoggers(fileLogger)
}

// initMultiWriter 初始化同时输出到控制台和文件的日志器
func initMultiWriter(logFilePath string) {

	if fileLogger == nil {
		fileLogger = &lumberjack.Logger{}
	}
	fileLogger.Filename = logFilePath
	fileLogger.MaxSize = globalLogger.logConf.MaxSize
	fileLogger.MaxBackups = globalLogger.logConf.MaxBackups
	fileLogger.MaxAge = globalLogger.logConf.KeepDays
	fileLogger.Compress = globalLogger.logConf.Compress

	// 创建一个同时写入控制台和文件的 Writer
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	globalLogger.output = multiWriter
	// 重新初始化所有日志器
	initLoggers(multiWriter)
}

// 设置方法 -----------------------------------------------------------------------
// SetUp 初始化日志记录器
func SetUp(logConf LogConf) error {
	mu.Lock()
	defer mu.Unlock()

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

	globalLogger.logConf = logConf
	globalLogger.logFlags = LogFlagsCommon
	globalLogger.hasRootFilePrefix = false
	globalLogger.logWriteStrategy = LoggingSync

	if globalLogger.logConf.Mode == "file" || globalLogger.logConf.Mode == "both" {
		if globalLogger.logConf.Path == "" {
			return errors.New("log path is required")
		}
	}

	// 设置编码
	switch logConf.Encoding {
	case LogEncodingPlain:
		globalLogger.encoder = &PlainEncoder{}
	case LogEncodingJSON:
		globalLogger.encoder = &JsonEncoder{}
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
	switch globalLogger.logConf.Mode {
	case "file":
		initFileLog(globalLogger.logConf.Path)
	case "both":
		initMultiWriter(globalLogger.logConf.Path)
	default:
		initLoggers(os.Stdout)
	}

	return nil
}

// DefaultLogConf 返回一个带有默认配置的 LogConf 实例
func DefaultLogConf() LogConf {
	return defaultLogConf
}

// SetupDefault 使用默认配置初始化日志记录器
func SetupDefault() error {
	err := SetUp(defaultLogConf)
	if err != nil {
		return fmt.Errorf("failed to set up default logger: %v", err)
	}
	return nil
}

// 判断是否是标准输出/错误流
func isStdStream(w io.Writer) bool {
	if w == os.Stdout || w == os.Stderr || w == io.Discard {
		return true
	}

	if f, ok := w.(*os.File); ok {
		return f == os.Stdout || f == os.Stderr
	}

	return false
}

// SetOutput 设置日志输出位置，自动更新Mode
func SetOutput(writer io.Writer) error {
	mu2.Lock()
	defer mu2.Unlock()

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	globalLogger.output = writer

	mode := LogModeConsole

	switch w := writer.(type) {
	case *os.File:
		// 文件输出
		if w.Name() == os.DevNull {
			mode = LogModeConsole // 特殊情况： /dev/null，仍视为console
					fmt.Println("console0")
		}else if isStdStream(w) {
			mode = LogModeConsole
					fmt.Println("console1")
		} else {
			mode = LogModeFile
			globalLogger.logConf.Path = w.Name()
		}
	default:
		// 尝试使用反射来检查是否为MultiWriter
		writerVal := reflect.ValueOf(writer)
		if writerVal.Kind() == reflect.Struct {
			if mw, ok := writer.(interface{ Writers() []io.Writer}); ok {
				hasFile := false
				hasConsole := false

				for _, wr := range mw.Writers() {
					if f, ok := wr.(*os.File); ok && !isStdStream(f) {
						hasFile = true
						globalLogger.logConf.Path = f.Name()
					}else if isStdStream(wr) {
						hasConsole = true
					fmt.Println("console2")
					}
				}

				if hasFile && hasConsole {
					mode = LogModeBoth
				} else if hasFile {
					mode = LogModeFile
					fmt.Println("file")
				} else {
					mode = LogModeConsole
					fmt.Println("console3")
				}	
			} else {

			}
		}

	}

	fmt.Println("mode：", mode)
	globalLogger.logConf.Mode = mode
	initLoggers(globalLogger.output)

	return nil
}

// 设置编码
func SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	mu.Lock()
	defer mu.Unlock()
	globalLogger.logConf.Encoding = encoding

	switch encoding {
	case LogEncodingPlain:
		globalLogger.encoder = &PlainEncoder{}
	case LogEncodingJSON:
		globalLogger.encoder = &JsonEncoder{}
	default:
		return fmt.Errorf("unsupported log encoding: %s", encoding)
	}
	return nil
}

// 设置日志文件最大大小
func SetMaxSize(maxSize int) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.logConf.MaxSize = maxSize

	// 重新初始化日志器以应用新设置
	if globalLogger.logConf.Mode == "file" {
		initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		initMultiWriter(globalLogger.logConf.Path)
	}
}

// 设置日志文件最大保留天数
func SetMaxAge(maxAge int) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.logConf.KeepDays = maxAge

	// 重新初始化日志器以应用新设置
	if globalLogger.logConf.Mode == "file" {
		initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		initMultiWriter(globalLogger.logConf.Path)
	}
}

// 设置日志文件最大保留数量
func SetMaxBackups(maxBackups int) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.logConf.MaxBackups = maxBackups

	if globalLogger.logConf.Mode == "file" {
		initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		initMultiWriter(globalLogger.logConf.Path)
	}
}

// 设置日志级别
func SetLogLevel(level LogLevel) error {
	mu.Lock()
	defer mu.Unlock()

	if level < LogLevelDebug || level > LogLevelPanic {
		return errors.New("invalid log level")
	}

	globalLogger.logConf.Level = int(level)
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
		globalLogger.hasRootFilePrefix = true
		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
	}

	globalLogger.logFlags = flags

	globalLogger.debugL.SetFlags(flags)
	globalLogger.infoL.SetFlags(flags)
	globalLogger.warnL.SetFlags(flags)
	globalLogger.errorL.SetFlags(flags)
	globalLogger.fatalL.SetFlags(flags)
	globalLogger.panicL.SetFlags(flags)
	return nil
}

// 设置日志同步还是异步
func SetLogWriteStrategy(strategy logWriteStrategy) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.logWriteStrategy = strategy
}

// 设置前缀
func SetPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.debugL.SetPrefix("[DEBUG] " + prefix)
	globalLogger.infoL.SetPrefix("[INFO] " + prefix)
	globalLogger.warnL.SetPrefix("[WARN] " + prefix)
	globalLogger.errorL.SetPrefix("[ERROR] " + prefix)
	globalLogger.fatalL.SetPrefix("[FATAL] " + prefix)
	globalLogger.panicL.SetPrefix("[PANIC] " + prefix)
}

func SetDebugPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.debugL.SetPrefix(prefix)
}

func SetDebugPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.debugL.SetPrefix("[DEBUG] " + prefix)
}

func SetInfoPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.infoL.SetPrefix(prefix)
}

func SetInfoPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.infoL.SetPrefix("[INFO] " + prefix)
}

func SetWarnPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.warnL.SetPrefix(prefix)
}

func SetWarnPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.warnL.SetPrefix("[WARN] " + prefix)
}

func SetErrorPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.errorL.SetPrefix(prefix)
}

func SetErrorPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.errorL.SetPrefix("[ERROR] " + prefix)
}

func SetFatalPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.fatalL.SetPrefix(prefix)
}

func SetFatalPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.fatalL.SetPrefix("[FATAL] " + prefix)
}

func SetPanicPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.panicL.SetPrefix(prefix)
}

func SetPanicPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.panicL.SetPrefix("[PANIC] " + prefix)
}

func SetPrefixWithoutDefaultPrefix(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger.debugL.SetPrefix(prefix)
	globalLogger.infoL.SetPrefix(prefix)
	globalLogger.warnL.SetPrefix(prefix)
	globalLogger.errorL.SetPrefix(prefix)
	globalLogger.fatalL.SetPrefix(prefix)
	globalLogger.panicL.SetPrefix(prefix)
}
