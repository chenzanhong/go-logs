package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	log "github.com/chenzanhong/logs/log_origin"
)

// findProjectRoot 查找项目的根目录（假设存在 go.mod 文件）
func findProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("无法获取当前文件信息")
	}
	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir { // 到达根目录
			break
		}
		dir = parentDir
	}

	return "", fmt.Errorf("未能找到项目根目录（go.mod 文件）")
}

// GetRelativePath 获取调用者的相对路径和行号
func GetRelativePath(skip int) (file string, line int) {
	projectRootOnce.Do(func() {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			projectRoot = "" // 如果找不到，则不使用相对路径
		}
	})

	_, path, line, _ := runtime.Caller(skip)
	relativePath, err := filepath.Rel(projectRoot, path)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return path, line
	}

	return relativePath, line
}

// GetRootfilePrefix 获取调用者的相对路径和行号
func GetRootfilePrefix(skip int) (rootfile string) {
	_, path, line, _ := runtime.Caller(skip)
	relativePath, err := filepath.Rel(projectRoot, path)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return fmt.Sprintf("%s %d: ", filepath.ToSlash(path), line)
	}

	return fmt.Sprintf("%s %d: ", filepath.ToSlash(relativePath), line)
}

// 根据日志级别获取对应的log.Logger实例
func getLoggerByLevel(logger *LogsLogger, level LogLevel) *log.Logger {
	switch level {
	case LogLevelDebug:
		return logger.debugL
	case LogLevelInfo:
		return logger.infoL
	case LogLevelWarn:
		return logger.warnL
	case LogLevelError:
		return logger.errorL
	case LogLevelFatal:
		return logger.fatalL
	case LogLevelPanic:
		return logger.panicL
	default:
		return nil
	}
}

// 判断字符串是否包含格式化符号
func containsFormatSpecifier(s string) bool {
	return regexp.MustCompile(`%(?:\.\*|\*[0-9]*|[0-9.]*[a-zA-Z])`).MatchString(s)
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

// LogLevel 转换为字符串
func LogLevelToString(level int) string {
	switch level {
	case int(LogLevelDebug):
		return "DEBUG"
	case int(LogLevelInfo):
		return "INFO"
	case int(LogLevelWarn):
		return "WARN"
	case int(LogLevelError):
		return "ERROR"
	case int(LogLevelFatal):
		return "FATAL"
	case int(LogLevelPanic):
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

func parseAndEncode(l *LogsLogger, skip int, v ...interface{}) string {
	if len(v) == 0 {
		return ""
	}

	var msg string = ""
	fields := make([]Field, 0, (len(v) + 7) / 2) // 预分配多3个字段空间，用于存储Level、timestamp和caller字段

	// Level字段
	fields = append(fields, Field{Key: "level", Value: LogLevelToString(l.logConf.Level)})
	// 日期时间字段
	fields = append(fields, Field{Key: "timestamp", Value: time.Now().Format(time.RFC3339)})
	// caller字段
	fields = append(fields, Field{Key: "caller", Value: GetRootfilePrefix(skip)})

	if len(v)%2 != 0 { // 如果参数个数为奇数，则第一个参数被视为消息
		msg = fmt.Sprint(v[0])
		fields = append(fields, Field{Key: "msg", Value: msg})
		v = v[1:]
	}

	for i := 0; i < len(v);  {
		key, ok := v[i].(string)
		if ok {
			fields = append(fields, Field{Key: key, Value: v[i+1]})
			i += 2 // 跳过键和值
			continue
		}
		key2, ok2 := v[i].(Field)
		if ok2 { // 检查是否为Field类型
			fields = append(fields, key2) // 直接添加Field类型
			i++ // 跳过Field类型
			continue
		}
		// 其他类型，非法
		break
	}
	encoder, _ := l.encoder.(StructuredEncoder) // 获取StructuredEncoder编码器实例

	return encoder.EncodeWithFieldsOrder(fields...) // 调用编码器方法进行编码（按照字段顺序）
}
var (
	timeFormat string = time.RFC3339
	useUTC     bool = true
)

func SetAsyncTimeFormat(format string, utc bool) {
	timeFormat = format
	useUTC = utc
}

func timeString() string {
	t := time.Now()
	if useUTC {
		t = t.UTC()
	}
	return t.Format(timeFormat)
}

func parseAndEncodeWithFields(l *LogsLogger, skip int, msg string, fields ...Field) string {
	newFields := make([]Field, 0, len(fields)+3) // 预分配多3个字段空间，用于存储Level、timestamp和caller字段

	// Level字段
	newFields = append(newFields, Field{Key: "level", Value: LogLevelToString(l.logConf.Level)})
	// 日期时间字段
	// newFields = append(newFields, Field{Key: "timestamp", Value: time.Now().Format(time.RFC3339)})
	newFields = append(newFields, Field{Key: "timestamp", Value: timeString()})
	// caller字段
	newFields = append(newFields, Field{Key: "caller", Value: GetRootfilePrefix(skip)})
	if msg != "" { // 如果msg不为空，则添加msg字段
		newFields = append(newFields, Field{Key: "msg", Value: msg})
	}

	newFields = append(newFields, fields...)

	encoder, _ := l.encoder.(StructuredEncoder) // 获取StructuredEncoder编码器实例

	return encoder.EncodeWithFieldsOrder(newFields...) // 调用编码器方法进行编码（按照字段顺序）
}