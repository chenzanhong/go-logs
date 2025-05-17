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

	log "github.com/chenzanhong/logs/log_origin"

	"gopkg.in/natefinch/lumberjack.v2"
)

func (l *LogsLogger) initLoggers(output io.Writer) error {

	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	l.output = output

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

	return nil
}

func (l *LogsLogger) initFileLog(logFilePath string) error {
	// 重新初始化所有日志器
	err := l.initLoggers(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    l.logConf.MaxSize,
		MaxBackups: l.logConf.MaxBackups,
		MaxAge:     l.logConf.KeepDays,
		Compress:   l.logConf.Compress,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize file logger: %v", err)
	}
	return nil
}

func (l *LogsLogger) initMultiWriter(logFilePath string) error {
	// 创建一个同时写入控制台和文件的 Writer
	multiWriter := io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    l.logConf.MaxSize,
		MaxBackups: l.logConf.MaxBackups,
		MaxAge:     l.logConf.KeepDays,
		Compress:   l.logConf.Compress,
	})

	// 重新初始化所有日志器
	err := l.initLoggers(multiWriter)
	if err != nil {
		return fmt.Errorf("failed to initialize multi writer: %v", err)
	}
	return nil
}

// 依据logConf和默认配置初始化日志器LogsLogger
func (l *LogsLogger) Setup(logConf LogConf) error {
	l.mu.Lock()
	defer l.mu.Unlock()

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

	l.logChan = make(chan *logItem, defaultLogChanSize)
	l.shutdownChan = make(chan struct{})
	l.itemPool = sync.Pool{
		New: func() interface{} {
			return &logItem{}
		},
	}
	// 初始化批量处理
	l.batchBuffer = make([][]byte, 0, batchSize)
	l.batchTicker = time.NewTicker(flushInterval)
	l.bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	// 重置l.closed和l.closeOnce
	if atomic.LoadInt32(&l.closed) == 1 {
		l.closeOnce = sync.Once{}
		atomic.StoreInt32(&l.closed, 0)
	}

	for i := 0; i < workerCount; i++ {
		l.wg.Add(1)
		go l.worker()
	}

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
	switch l.logConf.Mode {
	case "file":
		err = l.initFileLog(l.logConf.Path)
	case "both":
		err = l.initMultiWriter(l.logConf.Path)
	default:
		err = l.initLoggers(os.Stdout)
	}
	if err != nil {
		return fmt.Errorf("failed to set up logger: %v", err)
	}
	return nil
}

// SetupDefault 使用默认配置初始化日志记录器
func (l *LogsLogger) SetupDefault() error {
	err := l.Setup(defaultLogConf)
	if err != nil {
		return fmt.Errorf("failed to set up default logger: %v", err)
	}
	return nil
}

// SetOutput 设置日志输出位置，自动更新Mode
func (l *LogsLogger) SetOutput(writer io.Writer) error {
	l.mu.Lock()
	defer l.mu.Unlock()

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
				MaxSize:    l.logConf.MaxSize,
				MaxBackups: l.logConf.MaxBackups,
				MaxAge:     l.logConf.KeepDays,
				Compress:   l.logConf.Compress,
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

	l.output = writer
	l.logConf.Mode = mode
	l.logConf.Path = path
	l.initLoggers(l.output)

	return nil
}

// 设置编码
func (l *LogsLogger) SetEncoding(encoding string) error {
	// LogEncodingJSON、LOgEncodingPlain
	l.mu.Lock()
	defer l.mu.Unlock()
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
func (l *LogsLogger) SetMaxSize(maxSize int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logConf.MaxSize = maxSize

	var err error = nil
	// 重新初始化日志器以应用新设置
	if l.logConf.Mode == "file" {
		err = l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		err = l.initMultiWriter(l.logConf.Path)
	}

	if err != nil {
		return fmt.Errorf("failed to set max size: %v", err)
	}
	return nil
}

// 设置日志文件最大保留天数
func (l *LogsLogger) SetMaxAge(maxAge int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logConf.KeepDays = maxAge

	var err error = nil
	// 重新初始化日志器以应用新设置
	if l.logConf.Mode == "file" {
		err = l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		err = l.initMultiWriter(l.logConf.Path)
	}
	if err != nil {
		return fmt.Errorf("failed to set max age: %v", err)
	}
	return nil
}

// 设置日志文件最大保留数量
func (l *LogsLogger) SetMaxBackups(maxBackups int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logConf.MaxBackups = maxBackups

	var err error = nil
	// 重新初始化日志器以应用新设置
	if l.logConf.Mode == "file" {
		err = l.initFileLog(l.logConf.Path)
	} else if l.logConf.Mode == "both" {
		err = l.initMultiWriter(l.logConf.Path)
	}
	if err != nil {
		return fmt.Errorf("failed to set max backups: %v", err)
	}
	return nil
}

func (l *LogsLogger) SetLogLevel(level LogLevel) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < LogLevelDebug {
		return errors.New("invalid log level")
	}

	l.logConf.Level = int(level)
	return nil
}

// 设置标志
func (l *LogsLogger) SetFlags(flags int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

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
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logWriteStrategy = strategy
}

// 设置前缀
func (l *LogsLogger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugL.SetPrefix("[DEBUG] " + prefix)
	l.infoL.SetPrefix("[INFO] " + prefix)
	l.warnL.SetPrefix("[WARN] " + prefix)
	l.errorL.SetPrefix("[ERROR] " + prefix)
	l.fatalL.SetPrefix("[FATAL] " + prefix)
	l.panicL.SetPrefix("[PANIC] " + prefix)
}

func (l *LogsLogger) SetDebugPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugL.SetPrefix(prefix)
}

func (l *LogsLogger) SetDebugPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugL.SetPrefix("[DEBUG] " + prefix)
}

func (l *LogsLogger) SetInfoPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoL.SetPrefix(prefix)
}

func (l *LogsLogger) SetInfoPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoL.SetPrefix("[INFO] " + prefix)
}

func (l *LogsLogger) SetWarnPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnL.SetPrefix(prefix)
}

func (l *LogsLogger) SetWarnPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnL.SetPrefix("[WARN] " + prefix)
}

func (l *LogsLogger) SetErrorPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorL.SetPrefix(prefix)
}

func (l *LogsLogger) SetErrorPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorL.SetPrefix("[ERROR] " + prefix)
}

func (l *LogsLogger) SetFatalPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fatalL.SetPrefix(prefix)
}

func (l *LogsLogger) SetFatalPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fatalL.SetPrefix("[FATAL] " + prefix)
}

func (l *LogsLogger) SetPanicPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.panicL.SetPrefix(prefix)
}

func (l *LogsLogger) SetPanicPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.panicL.SetPrefix("[PANIC] " + prefix)
}

func (l *LogsLogger) SetPrefixWithoutDefaultPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugL.SetPrefix(prefix)
	l.infoL.SetPrefix(prefix)
	l.warnL.SetPrefix(prefix)
	l.errorL.SetPrefix(prefix)
	l.fatalL.SetPrefix(prefix)
	l.panicL.SetPrefix(prefix)
}
