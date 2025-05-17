package logs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 依据logConf和默认配置初始化日志器LogsLogger
func Setup(logConf LogConf) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

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
	globalLogger.logWriteStrategy = LoggingSync // 默认同步模式

	globalLogger.logChan = make(chan *logItem, defaultLogChanSize)
	globalLogger.shutdownChan = make(chan struct{})
	globalLogger.itemPool = sync.Pool{
		New: func() interface{} {
			return &logItem{}
		},
	}
	// 初始化批量处理
	globalLogger.batchBuffer = make([][]byte, 0, batchSize)
	globalLogger.batchTicker = time.NewTicker(flushInterval)
	globalLogger.bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	// 重置l.closed和l.closeOnce
	if atomic.LoadInt32(&globalLogger.closed) == 1 {
		globalLogger.closeOnce = sync.Once{}
		atomic.StoreInt32(&globalLogger.closed, 0)
	}

	for i := 0; i < workerCount; i++ {
		globalLogger.wg.Add(1)
		go globalLogger.worker()
	}

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

	// 获取项目根目录
	projectRootOnce.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	var err error = nil

	// 初始化输出
	switch globalLogger.logConf.Mode {
	case "file":
		err = globalLogger.initFileLog(globalLogger.logConf.Path)
	case "both":
		err = globalLogger.initMultiWriter(globalLogger.logConf.Path)
	default:
		err = globalLogger.initLoggers(os.Stdout)
	}
	if err != nil {
		return fmt.Errorf("failed to set up logger: %v", err)
	}
	return nil
}

// SetupDefault 使用默认配置初始化日志记录器
func SetupDefault() error {
	err := globalLogger.Setup(defaultLogConf)
	if err != nil {
		return fmt.Errorf("failed to set up default logger: %v", err)
	}
	return nil
}

// SetOutput 设置日志输出位置，自动更新Mode
func SetOutput(writer io.Writer) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	mode := LogModeConsole
	var path string

	switch w := writer.(type) {
	case *os.File:
		// 文件输出
		if w.Name() == os.DevNull {
			mode = LogModeConsole // 特殊情况： /dev/null，仍视为console
		} else if isStdStream(w) {
			mode = LogModeConsole
		} else {
			mode = LogModeFile
			path = w.Name()
			writer = &lumberjack.Logger{
				Filename:   path,
				MaxSize:    globalLogger.logConf.MaxSize,
				MaxBackups: globalLogger.logConf.MaxBackups,
				MaxAge:     globalLogger.logConf.KeepDays,
				Compress:   globalLogger.logConf.Compress,
			}
		}
	case *lumberjack.Logger:
		mode = LogModeFile
		path = w.Filename
	default:
		// 尝试使用反射来检查是否为MultiWriter
		writerVal := reflect.ValueOf(writer)
		if writerVal.Kind() == reflect.Struct {
			if mw, ok := writer.(interface{ Writers() []io.Writer }); ok {
				hasFile := false
				hasConsole := false

				for _, wr := range mw.Writers() {
					if f, ok := wr.(*os.File); ok && !isStdStream(f) {
						hasFile = true
						path = f.Name()
					} else if isStdStream(wr) {
						hasConsole = true
					} else if lj, ok := wr.(*lumberjack.Logger); ok {
						hasFile = true
						path = lj.Filename
					}
				}
				if hasFile && hasConsole {
					mode = LogModeBoth
				} else if hasFile {
					mode = LogModeFile
				} else {
					mode = LogModeConsole
				}
			}
		} else {
			fmt.Println("未知的writer类型，默认为console")
		}
	}

	globalLogger.output = writer
	globalLogger.logConf.Mode = mode
	globalLogger.logConf.Path = path
	globalLogger.initLoggers(globalLogger.output)

	return nil
}

// 设置编码
func SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
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
func SetMaxSize(maxSize int) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.logConf.MaxSize = maxSize

	var err error = nil
	// 重新初始化日志器以应用新设置
	if globalLogger.logConf.Mode == "file" {
		err = globalLogger.initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		err = globalLogger.initMultiWriter(globalLogger.logConf.Path)
	}

	if err != nil {
		return fmt.Errorf("failed to set max size: %v", err)
	}
	return nil
}

// 设置日志文件最大保留天数
func SetMaxAge(maxAge int) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.logConf.KeepDays = maxAge

	var err error = nil
	// 重新初始化日志器以应用新设置
	if globalLogger.logConf.Mode == "file" {
		err = globalLogger.initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		err = globalLogger.initMultiWriter(globalLogger.logConf.Path)
	}
	if err != nil {
		return fmt.Errorf("failed to set max age: %v", err)
	}
	return nil
}

// 设置日志文件最大保留数量
func SetMaxBackups(maxBackups int) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.logConf.MaxBackups = maxBackups

	var err error = nil
	// 重新初始化日志器以应用新设置
	if globalLogger.logConf.Mode == "file" {
		err = globalLogger.initFileLog(globalLogger.logConf.Path)
	} else if globalLogger.logConf.Mode == "both" {
		err = globalLogger.initMultiWriter(globalLogger.logConf.Path)
	}
	if err != nil {
		return fmt.Errorf("failed to set max backups: %v", err)
	}
	return nil
}

func SetLogLevel(level LogLevel) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	if level < LogLevelDebug {
		return errors.New("invalid log level")
	}

	globalLogger.logConf.Level = int(level)
	return nil
}

// 设置标志
func SetFlags(flags int) error {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

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

		// 检查并移除 Lshortfile 和 Llongfile，避免重复输出
		if flags&Lshortfile != 0 {
			flags = flags &^ Lshortfile
		}
		if flags&Llongfile != 0 {
			flags = flags &^ Llongfile
		}
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
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.logWriteStrategy = strategy
}

// 设置前缀
func SetPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.debugL.SetPrefix("[DEBUG] " + prefix)
	globalLogger.infoL.SetPrefix("[INFO] " + prefix)
	globalLogger.warnL.SetPrefix("[WARN] " + prefix)
	globalLogger.errorL.SetPrefix("[ERROR] " + prefix)
	globalLogger.fatalL.SetPrefix("[FATAL] " + prefix)
	globalLogger.panicL.SetPrefix("[PANIC] " + prefix)
}

func SetDebugPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.debugL.SetPrefix(prefix)
}

func SetDebugPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.debugL.SetPrefix("[DEBUG] " + prefix)
}

func SetInfoPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.infoL.SetPrefix(prefix)
}

func SetInfoPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.infoL.SetPrefix("[INFO] " + prefix)
}

func SetWarnPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.warnL.SetPrefix(prefix)
}

func SetWarnPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.warnL.SetPrefix("[WARN] " + prefix)
}

func SetErrorPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.errorL.SetPrefix(prefix)
}

func SetErrorPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.errorL.SetPrefix("[ERROR] " + prefix)
}

func SetFatalPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.fatalL.SetPrefix(prefix)
}

func SetFatalPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.fatalL.SetPrefix("[FATAL] " + prefix)
}

func SetPanicPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.panicL.SetPrefix(prefix)
}

func SetPanicPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.panicL.SetPrefix("[PANIC] " + prefix)
}

func SetPrefixWithoutDefaultPrefix(prefix string) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.debugL.SetPrefix(prefix)
	globalLogger.infoL.SetPrefix(prefix)
	globalLogger.warnL.SetPrefix(prefix)
	globalLogger.errorL.SetPrefix(prefix)
	globalLogger.fatalL.SetPrefix(prefix)
	globalLogger.panicL.SetPrefix(prefix)
}
