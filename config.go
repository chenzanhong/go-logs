package logs

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
)

const (
	// 日志标志
	Ldate          = 1 << iota                                     // 添加日期到输出
	Ltime                                                          // 添加时间到输出
	Lmicroseconds                                                  // 添加微秒到输出（覆盖 Ltime）
	Llongfile                                                      // 使用完整文件路径和行号
	Lshortfile                                                     // 使用短文件路径和行号（与 Llongfile 互斥）
	LUTC                                                           // 使用 UTC 时间格式
	Lmsgprefix                                                     // 将日志前缀放在每行日志的开头
	Lrootfile                                                      // 自定义的相对路径前缀
	LstdFlags      = Ldate | Ltime                                 // 标准日志标志：日期和时间
	LogFlagsCommon = Lmsgprefix | Ldate | Ltime | LUTC | Lrootfile // 示例：一个常见的标志组合
)
