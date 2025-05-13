package logs

// import (
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"runtime"
// 	"strings"
// 	"sync"

// 	"gopkg.in/natefinch/lumberjack.v2"
// )

// type LogConf struct {
// 	Mode       string `yaml:"mode"`        // 日志输出模式：console/file
// 	Level      int    `yaml:"level"`       // 日志级别：debug/info/warn/error/fatal/panic
// 	Encoding   string `yaml:"encoding"`    // 日志编码：plain/json
// 	Path       string `yaml:"path"`        // 日志文件路径（仅在文件模式下使用）
// 	MaxSize    int    `yaml:"max_size"`    // 日志文件最大大小（MB）
// 	MaxBackups int    `yaml:"max_backups"` // 日志文件最大保留数量
// 	KeepDays   int    `yaml:"keep_days"`   // 日志文件保留天数（仅在文件模式下使用）
// 	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件（仅在文件模式下使用）
// }

// type LogLevel int

// type LogsLogger struct { // 包含所有日志器的结构体
// 	debugL *log.Logger 
// 	infoL  *log.Logger
// 	warnL  *log.Logger
// 	errorL *log.Logger
// 	fatalL *log.Logger
// 	panicL *log.Logger

// 	hasRootFilePrefix bool // 是否打印自定义的相对路径前缀
// }

// type Encoder interface {
// 	Encode(v ...interface{}) string
// }

// type PlainEncoder struct{}

// func (e *PlainEncoder) Encode(v ...interface{}) string {
// 	return fmt.Sprint(v...)
// }

// type JsonEncoder struct{}

// func (e *JsonEncoder) Encode(v ...interface{}) string {
// 	b, err := json.Marshal(v)
// 	if err != nil {
// 		return fmt.Sprintf("JSON marshal error: %v", err)
// 	}
// 	return string(b)
// }

// const (
// 	// 日志级别
// 	LogLevelDebug LogLevel = iota
// 	LogLevelInfo
// 	LogLevelWarn
// 	LogLevelError
// 	LogLevelFatal
// 	LogLevelPanic
// )

// const (
// 	// 日志编码
// 	LogEncodingPlain = "plain" // 纯文本编码
// 	LogEncodingJSON  = "json"  // JSON 编码

// 	// 日志输出模式
// 	LogModeConsole = "console" // 输出到控制台
// 	LogModeFile    = "file"    // 输出到文件
// )

// const (
// 	// 日志标志
// 	Ldate          = 1 << iota                                     // 添加日期到输出
// 	Ltime                                                          // 添加时间到输出
// 	Lmicroseconds                                                  // 添加微秒到输出（覆盖 Ltime）
// 	Llongfile                                                      // 使用完整文件路径和行号
// 	Lshortfile                                                     // 使用短文件路径和行号（与 Llongfile 互斥）
// 	LUTC                                                           // 使用 UTC 时间格式
// 	Lmsgprefix                                                     // 将日志前缀放在每行日志的开头
// 	Lrootfile                                                      // 自定义的相对路径前缀
// 	LstdFlags      = Ldate | Ltime                                 // 标准日志标志：日期和时间
// 	LogFlagsCommon = Lmsgprefix | Ldate | Ltime | LUTC | Lrootfile // 示例：一个常见的标志组合
// )

//// 全局变量
// var (
// 	logConfig LogConf // 日志配置

// 	// 日志器
// 	debugLogger *log.Logger
// 	infoLogger  *log.Logger
// 	warnLogger  *log.Logger
// 	errorLogger *log.Logger
// 	fatalLogger *log.Logger
// 	panicLogger *log.Logger

// 	fileLogger *lumberjack.Logger // 用于文件输出的日志器

// 	encoder Encoder = &PlainEncoder{} // 默认使用 PlainEncoder

// 	logFlags = LogFlagsCommon // 默认日志标志

// 	currentLogLevel LogLevel = LogLevelInfo // 默认日志级别为 Info

// 	logOutput io.Writer = os.Stdout // 默认输出到控制台

// 	rootFilePrefix bool = false // 自定义的相对路径前缀
// 	projectRoot    string
// 	mu             sync.Mutex
// 	once           sync.Once
// )

// // InitLogger 初始化日志记录器
// // 可以根据需要调整日志级别和输出位置
// func initLoggers(output io.Writer) {
// 	flags := logFlags
// 	if flags&Lrootfile != 0 {
// 		rootFilePrefix = true
// 		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
// 	}
// 	debugLogger = log.New(output, "[DEBUG] ", flags)
// 	infoLogger = log.New(output, "[INFO] ", flags)
// 	warnLogger = log.New(output, "[WARN] ", flags)
// 	errorLogger = log.New(output, "[ERROR] ", flags)
// 	fatalLogger = log.New(os.Stderr, "[FATAL] ", flags)
// 	panicLogger = log.New(os.Stderr, "[PANIC] ", flags)
// }

// // initFileLog 初始化日志文件输出
// func initFileLog(logFilePath string) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	if fileLogger == nil {
// 		fileLogger = &lumberjack.Logger{}
// 	}
// 	fileLogger.Filename = logFilePath
// 	fileLogger.MaxSize = logConfig.MaxSize
// 	fileLogger.MaxBackups = logConfig.MaxBackups
// 	fileLogger.MaxAge = logConfig.KeepDays
// 	fileLogger.Compress = logConfig.Compress

// 	// 重新初始化所有日志器
// 	initLoggers(fileLogger)
// }

// // initMultiWriter 初始化同时输出到控制台和文件的日志器
// func initMultiWriter(logFilePath string) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	if fileLogger == nil {
// 		fileLogger = &lumberjack.Logger{}
// 	}
// 	fileLogger.Filename = logFilePath
// 	fileLogger.MaxSize = logConfig.MaxSize
// 	fileLogger.MaxBackups = logConfig.MaxBackups
// 	fileLogger.MaxAge = logConfig.KeepDays
// 	fileLogger.Compress = logConfig.Compress

// 	// 创建一个同时写入控制台和文件的 Writer
// 	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

// 	// 重新初始化所有日志器
// 	initLoggers(multiWriter)
// }

// // DefaultLogConf 返回一个带有默认配置的 LogConf 实例
// func DefaultLogConf() LogConf {
// 	return LogConf{
// 		Mode:       "console",         // 默认输出到控制台
// 		Level:      int(LogLevelInfo), // 默认日志级别为 Info
// 		Encoding:   "plain",           // 默认编码为 plain text
// 		Path:       "",                // 控制台模式下不需要路径
// 		MaxSize:    10,                // 每个日志文件最大 10MB（仅当 Mode 为 file 或 both 时有效）
// 		MaxBackups: 3,                 // 最多保留 3 个备份（仅当 Mode 为 file 或 both 时有效）
// 		KeepDays:   7,                 // 日志文件保留 7 天（仅当 Mode 为 file 或 both 时有效）
// 		Compress:   false,             // 压缩旧的日志文件（仅当 Mode 为 file 或 both 时有效）
// 	}
// }

// // 设置方法 -----------------------------------------------------------------------
// // SetUp 初始化日志记录器
// func SetUp(logConf LogConf) error {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	logConfig = logConf

// 	if logConfig.Mode == "file" || logConfig.Mode == "both" {
// 		if logConfig.Path == "" {
// 			return errors.New("log path is required")
// 		}
// 	}

// 	// 设置编码
// 	if err := SetEncoding(logConfig.Encoding); err != nil {
// 		log.Println("Failed to set encoding:", err)
// 		return fmt.Errorf("failed to set encoding: %v", err)
// 	}

// 	// 设置日志级别
// 	if err := SetLogLevel(LogLevel(logConf.Level)); err != nil {
// 		log.Println("Failed to set log level:", err)
// 		return fmt.Errorf("failed to set log level: %v", err)
// 	}

// 	// 获取项目根目录
// 	once.Do(func() {
// 		var err error
// 		projectRoot, err = findProjectRoot()
// 		if err != nil {
// 			projectRoot = "" // 如果找不到，则不使用相对路径
// 		}
// 	})

// 	// 初始化输出
// 	switch logConfig.Mode {
// 	case "file":
// 		initFileLog(logConfig.Path)
// 	case "both":
// 		initMultiWriter(logConfig.Path)
// 	default:
// 		initLoggers(os.Stdout)
// 	}

// 	return nil
// }

// // SetupDefault 使用默认配置初始化日志记录器
// func SetupDefault() error {
// 	defaultConfig := DefaultLogConf()
// 	err := SetUp(defaultConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to set up default logger: %v", err)
// 	}
// 	return nil
// }

// // SetOutput 设置日志输出位置
// func SetOutput(writer io.Writer) error {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	if writer == nil {
// 		return errors.New("writer cannot be nil")
// 	}

// 	switch logConfig.Mode {
// 	case "file":
// 		if fWriter, ok := writer.(*os.File); ok {
// 			initFileLog(fWriter.Name())
// 		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
// 			for _, w := range mw.Writers() {
// 				if fWriter, ok := w.(*os.File); ok {
// 					initFileLog(fWriter.Name())
// 					break
// 				}
// 			}
// 		} else {
// 			return errors.New("unsupported writer type for file mode")
// 		}
// 	case "both":
// 		// 如果是 both 模式，调用 initMultiWriter 并传递 writer 中的路径
// 		if fWriter, ok := writer.(*os.File); ok {
// 			initMultiWriter(fWriter.Name())
// 		} else if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
// 			// 如果是 MultiWriter，尝试从中找到 *os.File
// 			for _, w := range mw.Writers() {
// 				if fWriter, ok := w.(*os.File); ok {
// 					initMultiWriter(fWriter.Name())
// 					break
// 				}
// 			}
// 		} else {
// 			return errors.New("unsupported writer type for both mode")
// 		}
// 	case "console":
// 		// 如果是 console 模式，直接设置 logOutput
// 		logOutput = writer
// 		initLoggers(logOutput)
// 	default:
// 		return errors.New("unsupported log mode")
// 	}

// 	return nil
// }

// // 设置编码
// func SetEncoding(encoding string) error {
// 	// LogEncodingJSON、LOgEncodingPlain
// 	mu.Lock()
// 	defer mu.Unlock()
// 	logConfig.Encoding = encoding

// 	switch encoding {
// 	case LogEncodingPlain:
// 		encoder = &PlainEncoder{}
// 	case LogEncodingJSON:
// 		encoder = &JsonEncoder{}
// 	default:
// 		return fmt.Errorf("unsupported log encoding: %s", encoding)
// 	}
// 	return nil
// }

// // 设置日志文件最大大小
// func SetMaxSize(maxSize int) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	logConfig.MaxSize = maxSize

// 	// 重新初始化日志器以应用新设置
// 	if logConfig.Mode == "file" {
// 		initFileLog(logConfig.Path)
// 	} else if logConfig.Mode == "both" {
// 		initMultiWriter(logConfig.Path)
// 	}
// }

// // 设置日志文件最大保留天数
// func SetMaxAge(maxAge int) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	logConfig.KeepDays = maxAge

// 	// 重新初始化日志器以应用新设置
// 	if logConfig.Mode == "file" {
// 		initFileLog(logConfig.Path)
// 	} else if logConfig.Mode == "both" {
// 		initMultiWriter(logConfig.Path)
// 	}
// }

// // 设置日志文件最大保留数量
// func SetMaxBackups(maxBackups int) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	logConfig.MaxBackups = maxBackups

// 	if logConfig.Mode == "file" {
// 		initFileLog(logConfig.Path)
// 	} else if logConfig.Mode == "both" {
// 		initMultiWriter(logConfig.Path)
// 	}
// }

// func SetLogLevel(level LogLevel) error {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	if level < LogLevelDebug {
// 		log.Printf("Invalid log level")
// 		return errors.New("Invalid log level")
// 	}

// 	currentLogLevel = level
// 	return nil
// }

// // 设置标志
// func SetFlags(flags int) error {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	// 对flags的合法性进行检查
// 	// 检查是否设置了无效的标志
// 	const vaildFlags = Ldate | Ltime | Lmicroseconds | Llongfile | Lshortfile | LUTC | Lmsgprefix | Lrootfile
// 	if flags < 0 || (flags & ^vaildFlags) != 0 {
// 		log.Println("Invalid flags value")
// 		return errors.New("Invalid flags value")
// 	}

// 	// 检查是否设置了 Ldate、Ltime 或 Lmicroseconds 标志
// 	if flags&(Ldate|Ltime|Lmicroseconds) == 0 {
// 		// 如果没有设置日期、时间或微秒，设置默认的 Ldate | Ltime
// 		flags = Ldate | Ltime
// 	}

// // 	// 检查是否设置了 Lrootfile 标志
// 	if flags&Lrootfile != 0 {
// 		rootFilePrefix = true
// 		flags = flags &^ Lrootfile // 移除 Lrootfile 标志
// 	}

// 	logFlags = flags

// 	debugLogger.SetFlags(flags)
// 	infoLogger.SetFlags(flags)
// 	warnLogger.SetFlags(flags)
// 	errorLogger.SetFlags(flags)
// 	fatalLogger.SetFlags(flags)
// 	panicLogger.SetFlags(flags)
// 	return nil
// }

// // 设置前缀
// func SetPrefix(prefix string) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	debugLogger.SetPrefix(prefix)
// 	infoLogger.SetPrefix(prefix)
// 	warnLogger.SetPrefix(prefix)
// 	errorLogger.SetPrefix(prefix)
// 	fatalLogger.SetPrefix(prefix)
// 	panicLogger.SetPrefix(prefix)
// }
// func SetDebugPrefix(prefix string) {
// 	debugLogger.SetPrefix(prefix)
// }
// func SetInfoPrefix(prefix string) {
// 	infoLogger.SetPrefix(prefix)
// }
// func SetWarnPrefix(prefix string) {
// 	warnLogger.SetPrefix(prefix)
// }
// func SetErrorPrefix(prefix string) {
// 	errorLogger.SetPrefix(prefix)
// }
// func SetFatalPrefix(prefix string) {
// 	fatalLogger.SetPrefix(prefix)
// }
// func SetPanicPrefix(prefix string) {
// 	panicLogger.SetPrefix(prefix)
// }
// func SetLoggerPrefix(logger *log.Logger, newPrefix string) {
// 	logger.SetPrefix(newPrefix)
// }

// // --------------------------------------------------------------------------------
// // findProjectRoot 查找项目的根目录（假设存在 go.mod 文件）
// func findProjectRoot() (string, error) {
// 	_, filename, _, ok := runtime.Caller(0)
// 	if !ok {
// 		return "", fmt.Errorf("无法获取当前文件信息")
// 	}
// 	dir := filepath.Dir(filename)

// 	for {
// 		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
// 			return dir, nil
// 		}
// 		parentDir := filepath.Dir(dir)
// 		if parentDir == dir { // 到达根目录
// 			break
// 		}
// 		dir = parentDir
// 	}

// 	return "", fmt.Errorf("未能找到项目根目录（go.mod 文件）")
// }

// // GetRelativePath 获取调用者的相对路径和行号
// func GetRelativePath(skip int) (file string, line int) {
// 	once.Do(func() {
// 		var err error
// 		projectRoot, err = findProjectRoot()
// 		if err != nil {
// 			projectRoot = "" // 如果找不到，则不使用相对路径
// 		}
// 	})

// 	_, path, line, _ := runtime.Caller(skip)
// 	relativePath, err := filepath.Rel(projectRoot, path)
// 	if err != nil || strings.HasPrefix(relativePath, "..") {
// 		return path, line
// 	}

// 	return relativePath, line
// }

// func GetLogPrefix(skip int) (logPrefix string) {
// 	_, path, line, _ := runtime.Caller(skip)
// 	relativePath, err := filepath.Rel(projectRoot, path)
// 	if err != nil || strings.HasPrefix(relativePath, "..") {
// 		return fmt.Sprintf("%s %d: ", path, line)
// 	}

// 	return fmt.Sprintf("%s %d: ", relativePath, line)
// }

// func outputLog(logger *log.Logger, skip int, v ...interface{}) {
// 	if currentLogLevel > LogLevel(logConfig.Level) {
// 		return
// 	}
// 	msg := encoder.Encode(v...)
// 	if rootFilePrefix {
// 		msg = GetLogPrefix(skip) + msg
// 	}
// 	logger.Output(skip, msg)
// }

// func outputLogf(logger *log.Logger, skip int, format string, v ...interface{}) {
// 	if currentLogLevel > LogLevel(logConfig.Level) {
// 		return
// 	}
// 	msg := encoder.Encode(fmt.Sprintf(format, v...))
// 	if rootFilePrefix {
// 		msg = GetLogPrefix(skip) + msg
// 	}
// 	logger.Output(skip, msg)
// }

// // Info 输出 INFO 日志
// func Info(v ...interface{}) {
// 	outputLog(infoLogger, 3, v...)
// }

// func Infof(format string, v ...interface{}) {
// 	outputLogf(infoLogger, 3, format, v...)
// }

// // Warn 输出 WARN 日志
// func Warn(v ...interface{}) {
// 	outputLog(warnLogger, 3, v...)
// }

// func Warnf(format string, v ...interface{}) {
// 	outputLogf(warnLogger, 3, format, v...)
// }

// // Error 输出 ERROR 日志
// func Error(v ...interface{}) {
// 	outputLog(errorLogger, 3, v...)
// }

// func Errorf(format string, v ...interface{}) {
// 	outputLogf(errorLogger, 3, format, v...)
// }

// // Fatal 输出 FATAL 日志并退出程序
// func Fatal(v ...interface{}) {
// 	outputLog(fatalLogger, 3, v...)
// 	os.Exit(1)
// }

// func Fatalf(format string, v ...interface{}) {
// 	outputLogf(fatalLogger, 3, format, v...)
// 	os.Exit(1)
// }

// // Panic 输出 PANIC 日志并触发 panic
// func Panic(v ...interface{}) {
// 	outputLog(panicLogger, 3, v...)
// 	panic(fmt.Sprint(v...))
// }

// func Panicf(format string, v ...interface{}) {
// 	outputLogf(panicLogger, 3, format, v...)
// 	panic(fmt.Sprintf(format, v...))
// }

// func Debug(v ...interface{}) {
// 	outputLog(debugLogger, 3, v...)
// }

// func Debugf(format string, v ...interface{}) {
// 	outputLogf(debugLogger, 3, format, v...)
// }


// // LogsLogger 的 output 方法
// func (l *LogsLogger) Debug(v ...interface{}) {
// 	outputLog(l.debugL, l.hasRootFilePrefix, 3, v...)
// }
// func (l *LogsLogger) Debugf(format string, v ...interface{}) {
// 	outputLogf(l.debugL, l.hasRootFilePrefix, 3, format, v...)
// }

// func (l *LogsLogger) Info(v ...interface{}) {
// 	outputLog(l.infoL, l.hasRootFilePrefix, 3, v...)
// }
// func (l *LogsLogger) Infof(format string, v ...interface{}) {
// 	outputLogf(l.infoL, l.hasRootFilePrefix, 3, format, v...)
// }

// func (l *LogsLogger) Warn(v ...interface{}) {
// 	outputLog(l.warnL, l.hasRootFilePrefix, 3, v...)
// }
// func (l *LogsLogger) Warnf(format string, v ...interface{}) {
// 	outputLogf(l.warnL, l.hasRootFilePrefix, 3, format, v...)
// }

// func (l *LogsLogger) Error(v ...interface{}) {
// 	outputLog(l.errorL, l.hasRootFilePrefix, 3, v...)
// }
// func (l *LogsLogger) Errorf(format string, v ...interface{}) {
// 	outputLogf(l.errorL, l.hasRootFilePrefix, 3, format, v...)
// }

// func (l *LogsLogger) Fatal(v ...interface{}) {
// 	outputLog(l.fatalL, l.hasRootFilePrefix, 3, v...)
// 	os.Exit(1)
// }
// func (l *LogsLogger) Fatalf(format string, v ...interface{}) {
// 	outputLogf(l.fatalL, l.hasRootFilePrefix, 3, format, v...)
// 	os.Exit(1)
// }

// func (l *LogsLogger) Panic(v ...interface{}) {
// 	outputLog(l.panicL, l.hasRootFilePrefix, 3, v...)
// 	panic(fmt.Sprint(v...))
// }

// func (l *LogsLogger) Panicf(format string, v ...interface{}) {
// 	outputLogf(l.panicL, l.hasRootFilePrefix, 3, format, v...)
// 	panic(fmt.Sprintf(format, v...))
// }


// func NewLogConfWithParams(mode string, level LogLevel, encoding string, path string, maxSize int, maxBackups int, keepDays int, compress bool) LogConf {
// 	return LogConf{
// 		Mode:       mode,
// 		Level:      int(level),
// 		Encoding:   encoding,
// 		Path:       path,
// 		MaxSize:    maxSize,
// 		MaxBackups: maxBackups,
// 		KeepDays:   keepDays,
// 		Compress:   compress,
// 	}
// }

// func NewLogConfWithDefaults(custom LogConf) LogConf {
// 	// 从默认配置开始
// 	conf := defaultLogConf

// 	// 如果用户提供了特定的值，则覆盖默认值
// 	if custom.Mode != "" {
// 		conf.Mode = custom.Mode
// 	}
// 	if custom.Level != 0 { // 注意：0 是 LogLevelInfo 的默认值，确保你的逻辑正确处理这种情况
// 		conf.Level = custom.Level
// 	}
// 	if custom.Encoding != "" {
// 		conf.Encoding = custom.Encoding
// 	}
// 	if custom.Path != "" {
// 		conf.Path = custom.Path
// 	}
// 	if custom.MaxSize != 0 {
// 		conf.MaxSize = custom.MaxSize
// 	}
// 	if custom.MaxBackups != 0 {
// 		conf.MaxBackups = custom.MaxBackups
// 	}
// 	if custom.KeepDays != 0 {
// 		conf.KeepDays = custom.KeepDays
// 	}
// 	if custom.Compress {
// 		conf.Compress = custom.Compress
// 	}

// 	return conf
// }

// func NewDefaultLogger(flag int) *LogsLogger { // console 输出
// 	var logger *LogsLogger = &LogsLogger{}
// 	if flag&Lrootfile != 0 {
// 		logger.hasRootFilePrefix = true
// 		flag = flag &^ Lrootfile // 移除 Lrootfile 标志
// 	}
// 	logger.debugL = log.New(os.Stdout, "DEBUG: ", flag)
// 	logger.infoL = log.New(os.Stdout, "INFO: ", flag)
// 	logger.warnL = log.New(os.Stdout, "WARN: ", flag)
// 	logger.errorL = log.New(os.Stdout, "ERROR: ", flag)
// 	logger.fatalL = log.New(os.Stderr, "FATAL: ", flag)
// 	logger.panicL = log.New(os.Stderr, "PANIC: ", flag)

// 	return logger
// }

// func newLogger(writer io.Writer, flag int, prefixFormat string) (*LogsLogger, error) {
// 	if writer == nil {
// 		return nil, errors.New("writer cannot be nil")
// 	}

// 	logger := &LogsLogger{}

// 	if flag&Lrootfile != 0 {
// 		logger.hasRootFilePrefix = true
// 		flag = flag &^ Lrootfile // 移除 Lrootfile 标志
// 	}

// 	// prefixes := map[*log.Logger]string{
// 	// 	logger.Debug: "DEBUG: ",
// 	// 	logger.Info:  "INFO: ",
// 	// 	logger.Warn:  "WARN: ",
// 	// 	logger.Error: "ERROR: ",
// 	// 	logger.Fatal: "FATAL: ",
// 	// 	logger.Panic: "PANIC: ",
// 	// }

// 	// // 可以自定义每个级别的前缀格式
// 	// if prefixFormat == "json" {
// 	// 	// 如果是 JSON 格式，可以在这里做定制化处理
// 	// }

// 	// 初始化每个级别的日志器
// 	logger.debugL = log.New(writer, "DEBUG: ", flag)
// 	logger.infoL = log.New(writer, "INFO: ", flag)
// 	logger.warnL = log.New(writer, "WARN: ", flag)
// 	logger.errorL = log.New(writer, "ERROR: ", flag)
// 	logger.fatalL = log.New(writer, "FATAL: ", flag)
// 	logger.panicL = log.New(writer, "PANIC: ", flag)

// 	return logger, nil
// }

// func NewFileLogger(filename string, flag int) (*LogsLogger, error) {
// 	if filename == "" {
// 		return nil, errors.New("filename cannot be empty")
// 	}

// 	writer := &lumberjack.Logger{
// 		Filename:   filename,
// 		MaxSize:    defaultLogConf.MaxSize,
// 		MaxBackups: defaultLogConf.MaxBackups,
// 		MaxAge:     defaultLogConf.KeepDays,
// 		Compress:   defaultLogConf.Compress,
// 	}

// 	return newLogger(writer, flag, defaultLogConf.Encoding)
// }

// func NewMultiWriterLogger(filename string, flag int) (*LogsLogger, error) {
// 	if filename == "" {
// 		return nil, errors.New("filename cannot be empty")
// 	}

// 	fileWriter := &lumberjack.Logger{
// 		Filename:   filename,
// 		MaxSize:    defaultLogConf.MaxSize,
// 		MaxBackups: defaultLogConf.MaxBackups,
// 		MaxAge:     defaultLogConf.KeepDays,
// 		Compress:   defaultLogConf.Compress,
// 	}

// 	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

// 	return newLogger(multiWriter, flag, defaultLogConf.Encoding)
// }
