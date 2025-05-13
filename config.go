package logs

import (
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConf struct {
	Mode       string `yaml:"mode"`        // 日志输出模式：console/file
	Level      int    `yaml:"level"`       // 日志级别：debug/info/warn/error/fatal/panic
	Encoding   string `yaml:"encoding"`    // 日志编码：plain/json
	Path       string `yaml:"path"`        // 日志文件路径（仅在文件模式下使用）
	MaxSize    int    `yaml:"max_size"`    // 日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 日志文件最大保留数量
	KeepDays   int    `yaml:"keep_days"`   // 日志文件保留天数（仅在文件模式下使用）
	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件（仅在文件模式下使用）
}

type LogLevel int

type LogsLogger struct { // 包含所有日志器的结构体
	debugL *log.Logger
	infoL  *log.Logger
	warnL  *log.Logger
	errorL *log.Logger
	fatalL *log.Logger
	panicL *log.Logger

	hasRootFilePrefix bool // 是否打印自定义的相对路径前缀
	output            io.Writer
	logFlags          int
	encoder           Encoder // 编码器
	logConf           LogConf // 日志配置
}

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelPanic
)

const (
	LogEncodingPlain = "plain" // 纯文本编码
	LogEncodingJSON  = "json"  // JSON 编码

	LogModeConsole = "console" // 输出到控制台
	LogModeFile    = "file"    // 输出到文件
	LogModeBoth    = "both"    // 同时输出到控制台和文件
)

const (
	// 日志标志
	Ldate          = 1 << iota                              // 添加日期到输出
	Ltime                                                   // 添加时间到输出
	Lmicroseconds                                           // 添加微秒到输出（覆盖 Ltime）
	Llongfile                                               // 使用完整文件路径和行号
	Lshortfile                                              // 使用短文件路径和行号（与 Llongfile 互斥）
	LUTC                                                    // 使用 UTC 时间格式
	Lmsgprefix                                              // 将日志前缀放在每行日志的开头
	Lrootfile                                               // 自定义的相对路径前缀
	LstdFlags      = Ldate | Ltime                          // 标准日志标志：日期和时间
	LogFlagsCommon = Lmsgprefix | Ldate | Ltime | Lrootfile // 示例：一个常见的标志组合
)

var (
	// logConfig      LogConf    // 日志配置
	defaultLogConf = LogConf{ // 默认日志配置
		Mode:       "console",         // 默认输出到控制台,
		Level:      int(LogLevelInfo), // 默认日志级别为 Info
		Encoding:   "plain",           // 默认编码为 plain text
		Path:       "",                // 控制台模式下不需要路径
		MaxSize:    10,                // 默认每个日志文件最大 10MB
		MaxBackups: 3,                 // 默认最多保留 3 个备份
		KeepDays:   7,                 // 默认日志文件保留 7 天
		Compress:   false,             // 默认不压缩旧的日志文件
	}

	defaultLogger = &LogsLogger{ // 默认日志器
		debugL:            log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		infoL:             log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warnL:             log.New(os.Stdout, "WARN: ", log.LstdFlags),
		errorL:            log.New(os.Stdout, "ERROR: ", log.LstdFlags),
		fatalL:            log.New(os.Stderr, "FATAL: ", log.LstdFlags),
		panicL:            log.New(os.Stderr, "PANIC: ", log.LstdFlags),
		encoder:           &PlainEncoder{},
		output:            os.Stdout,
		logFlags:          LogFlagsCommon,
		hasRootFilePrefix: false,
		logConf:           defaultLogConf,
	}

	globalLogger = &LogsLogger{} // 全局日志器

	// // 日志器
	// debugLogger *log.Logger
	// infoLogger  *log.Logger
	// warnLogger  *log.Logger
	// errorLogger *log.Logger
	// fatalLogger *log.Logger
	// panicLogger *log.Logger

	fileLogger *lumberjack.Logger // 用于文件输出的日志器

	// encoder Encoder = &PlainEncoder{} // 默认使用 PlainEncoder

	// logFlags = LogFlagsCommon // 默认日志标志

	currentLogLevel LogLevel = LogLevelInfo // 默认日志级别为 Info

	// logOutput io.Writer = os.Stdout // 默认输出到控制台

	// rootFilePrefix bool = false // 自定义的相对路径前缀
	projectRoot string
	mu          sync.Mutex
	once        sync.Once
)
