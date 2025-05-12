package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConf struct {
	Mode       string `yaml:"mode"`        // 日志输出模式：console/file
	Encoding   string `yaml:"encoding"`    // 日志编码：plain/json
	Path       string `yaml:"path"`        // 日志文件路径（仅在文件模式下使用）
	MaxSize    int    `yaml:"max_size"`    // 日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 日志文件最大保留数量
	KeepDays   int    `yaml:"keep_days"`   // 日志文件保留天数（仅在文件模式下使用）
	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件（仅在文件模式下使用）
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

const (
	Ldate          = log.Ldate                                                      // 添加日期到输出
	Ltime          = log.Ltime                                                      // 添加时间到输出
	Lmicroseconds  = log.Lmicroseconds                                              // 添加微秒到输出（覆盖 Ltime）
	Llongfile      = log.Llongfile                                                  // 使用完整文件路径和行号
	Lshortfile     = log.Lshortfile                                                 // 使用短文件路径和行号（与 Llongfile 互斥）
	LUTC           = log.LUTC                                                       // 使用 UTC 时间格式
	Lmsgprefix     = log.Lmsgprefix                                                 // 将日志前缀放在每行日志的开头
	LstdFlags      = log.LstdFlags                                                  // 等价于 Ldate | Ltime
	LogFlagsCommon = Ldate | Ltime | Lmicroseconds | LUTC | Lmsgprefix | Lshortfile // 示例：一个常见的标志组合
)

// 全局变量
var (
	logConfig LogConf

	// 日志器
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	panicLogger *log.Logger

	fileLogger *lumberjack.Logger

	logFlags = LogFlagsCommon

	logOutput io.Writer = os.Stdout // 默认输出到控制台
	mu        sync.Mutex
)

// InitLogger 初始化日志记录器
// 可以根据需要调整日志级别和输出位置
func initLoggers(output io.Writer) {

	debugLogger = log.New(output, "[DEBUG]", logFlags)
	infoLogger = log.New(output, "[INFO] ", logFlags)
	warnLogger = log.New(output, "[WARN] ", logFlags)
	errorLogger = log.New(output, "[ERROR] ", logFlags)
	fatalLogger = log.New(os.Stderr, "[FATAL] ", logFlags)
	panicLogger = log.New(os.Stderr, "[PANIC] ", logFlags)
}

// initFileLog 初始化日志文件输出（可选）
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

	logOutput = fileLogger

	if logConfig.Mode == "both" {
		// 设置多路输出：控制台 + 文件
		logOutput = io.MultiWriter(os.Stdout, fileLogger)
	}

	// 重新初始化所有日志器
	initLoggers(logOutput)
}

// DefaultLogConf 返回一个带有默认配置的 LogConf 实例
func DefaultLogConf() LogConf {
	return LogConf{
		Mode:       "console", // 默认输出到控制台
		Encoding:   "plain",   // 默认编码为 plain text
		Path:       "",        // 控制台模式下不需要路径
		MaxSize:    10,        // 每个日志文件最大 10MB（仅当 Mode 为 file 或 both 时有效）
		MaxBackups: 3,         // 最多保留 3 个备份（仅当 Mode 为 file 或 both 时有效）
		KeepDays:   7,         // 日志文件保留 7 天（仅当 Mode 为 file 或 both 时有效）
		Compress:   false,     // 压缩旧的日志文件（仅当 Mode 为 file 或 both 时有效）
	}
}

// 设置方法 -----------------------------------------------------------------------
// SetUp 初始化日志记录器
func SetUp(logConf LogConf) {
	mu.Lock()
	defer mu.Unlock()

	logConfig = logConf

	switch logConfig.Mode {
	case "file", "both":
		if logConfig.Path == "" {
			log.Fatal("log path is required when mode is 'file' or 'both'")
		}
		initFileLog(logConfig.Path)
	default:
		initLoggers(os.Stdout)
	}
}

// SetupDefault 使用默认配置初始化日志记录器
func SetupDefault() {
	defaultConfig := DefaultLogConf()
	SetUp(defaultConfig)
}

// SetOutput 设置日志输出位置（可选）
func SetOutput(writer io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if writer == nil {
		log.Fatal("writer cannot be nil")
	}

	// 判断当前的Mode
	if logConfig.Mode == "file" || logConfig.Mode == "both" {
		// 如果是文件输出，需要更新所有日志器
		if _, ok := writer.(*os.File); ok {
			// 如果是文件输出，需要更新所有日志器
			initLoggers(writer)
		}
	} else if logConfig.Mode == "console" {
		// 如果是控制台输出，只需要更新全局输出
		logOutput = writer
		initLoggers(writer)
	} else {
		log.Fatal("unsupported log mode")
	}
}

// 设置日志文件最大大小（可选）
func SetMaxSize(maxSize int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.MaxSize = maxSize

	// 重新初始化日志器以应用新设置
	if logConfig.Mode == "file" || logConfig.Mode == "both" {
		initFileLog(logConfig.Path)
	}
}

// 设置日志文件最大保留天数（可选）
func SetMaxAge(maxAge int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.KeepDays = maxAge

	// 重新初始化日志器以应用新设置
	if logConfig.Mode == "file" || logConfig.Mode == "both" {
		initFileLog(logConfig.Path)
	}
}

// 设置日志文件最大保留数量（可选）
func SetMaxBackups(maxBackups int) {
	mu.Lock()
	defer mu.Unlock()
	logConfig.MaxBackups = maxBackups

	if logConfig.Mode == "file" || logConfig.Mode == "both" {
		initFileLog(logConfig.Path)
	}
}

// 设置标志
func SetFlags(flags int) {
	mu.Lock()
	defer mu.Unlock()

	logFlags = flags

	initLoggers(logOutput)
}

// --------------------------------------------------------------------------------
// GetRelativePath 获取调用者的相对路径和行号
func GetRelativePath() (file string, line int) {
	_, filepath, line, _ := runtime.Caller(0)
	i := strings.Index(filepath, "/server/")
	if i != -1 {
		filepath = "/" + filepath[i+len("/server/"):] // 加1是为了跳过"/"
	} else {
		filepath = "" // 或者其他默认值/错误处理
	}
	return filepath, line
}

func GetLogPrefix(skip int) (logPrefix string) {
	_, filepath, line, _ := runtime.Caller(skip) // 一般为2
	i := strings.Index(filepath, "/server/")
	if i != -1 {
		filepath = "/" + filepath[i+len("/server/"):]
	} else {
		filepath = "" // 或者其他默认值/错误处理
	}
	return fmt.Sprintf("%s %d: ", filepath, line)
}

// Info 输出 INFO 日志
func Info(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	infoLogger.Output(2, msg)
}

func Infof(format string, v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprintf(format, v...)
	infoLogger.Output(2, msg)
}

// Warn 输出 WARN 日志
func Warn(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	warnLogger.Output(2, msg)
}

func Warnf(format string, v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprintf(format, v...)
	warnLogger.Output(2, msg)
}

// Error 输出 ERROR 日志
func Error(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	errorLogger.Output(2, msg)
}

func Errorf(format string, v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprintf(format, v...)
	errorLogger.Output(2, msg)
}

// Fatal 输出 FATAL 日志并退出程序
func Fatal(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	fatalLogger.Output(2, msg)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprintf(format, v...)
	fatalLogger.Output(2, msg)
	os.Exit(1)
}

// Panic 输出 PANIC 日志并触发 panic
func Panic(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	panicLogger.Output(2, msg)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprintf(format, v...)
	panicLogger.Output(2, msg)
	panic(fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	msg := GetLogPrefix(2) + fmt.Sprint(v...)
	debugLogger.Output(2, msg)
}
