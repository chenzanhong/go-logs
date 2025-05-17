package logs

import (
	"fmt"
	"os"
	"sync/atomic"
)

/*
outputLogf 方法的作用是将格式化后的日志消息输出到日志器的输出流中。
该方法接收一个日志级别、调用栈深度和格式化字符串，以及可变数量的参数。
在方法内部，首先使用日志器的编码器将参数格式化得到消息。
然后，根据日志器的日志级别，获取相应级别的内部日志器。
如果日志器的日志写入策略是 LoggingSync（同步写入），则直接调用内部日志器的 Output 方法将消息输出到日志器的输出流中。
如果日志器的日志写入策略是 LoggingAsync（异步写入），则将日志项放入日志通道中，由工作协程异步处理。
在工作协程中，从日志通道中取出日志项，然后调用内部日志器的 Output 方法将消息输出到日志器的输出流中。
最后，将日志项放回对象池中，以便后续重复使用。
*/
func (l *LogsLogger) outputLogf(level LogLevel, skip int, format string, v ...interface{}) {
	// var msg string
	// msg = l.encoder.Encode(v...) // 格式化消息，PlainEncoder.Encode方法只是单纯拼接，不会格式化
	msg := fmt.Sprintf(format, v...) // 格式化消息

	item := l.itemPool.Get().(*logItem)
	item.logger = l
	item.level = level
	item.skip = skip + 1

	if l.hasRootFilePrefix { // 是否打印自定义的相对路径前缀
		msg = GetRootfilePrefix(skip) + msg
	}

	item.msg = msg

	if l.logWriteStrategy == LoggingSync { // 同步写入
		internalLogger := getLoggerByLevel(l, level) // 获取相应级别的内部日志器
		internalLogger.Output(skip, msg)             // 输出日志（log.Logger.Output）
		l.itemPool.Put(item)                         // 放回对象池
	} else { // 异步写入
		if atomic.LoadInt32(&l.closed) == 1 { // 检查是否已关闭
			l.itemPool.Put(item) // 放回对象池
			return
		}
		select {
		case l.logChan <- item:
		default:
			fmt.Printf("日志通道已满，同步写入该日志: %s", msg)
			internalLogger := getLoggerByLevel(l, level)
			internalLogger.Output(skip, msg)
			l.itemPool.Put(item)
		}
	}
}

/*
outputLog 方法的作用是将日志消息输出到日志器的输出流中。
该方法接收一个日志级别、调用栈深度和可变数量的参数。
在方法内部，首先检查日志器的编码器是否实现了 StructuredEncoder 接口。
如果实现了 StructuredEncoder 接口，则表示使用 JSON 编码。
在这种情况下，首先使用 parseAndEncodeWithFields 方法解析参数得到消息，并将其放入日志项中。
然后，将日志项放入日志通道中，由工作协程异步处理。
如果编码器没有实现 StructuredEncoder 接口，则表示使用 Plain 编码。
在这种情况下，首先使用编码器将参数格式化得到消息。
然后，根据日志器的日志级别，获取相应级别的内部日志器。
如果日志器的日志写入策略是 LoggingSync（同步写入），则直接调用内部日志器的 Output 方法将消息输出到日志器的输出流中。
如果日志器的日志写入策略是 LoggingAsync（异步写入），则将日志项放入日志通道中，由工作协程异步处理。
在工作协程中，从日志通道中取出日志项，然后调用内部日志器的 Output 方法将消息输出到日志器的输出流中。
最后，将日志项放回对象池中，以便后续重复使用。

JSON 编码:
默认包含Level、，timestamp、caller、msg字段，
其他字段通过parseAndEncodeWithFields解析得到，并按输入顺序添加到msg中。
示例，通过parseAndEncodeWithFields得到的msg：

	{
		"level": "INFO",
		"timestamp": "2025/05/16 13:27:40",
		"caller": "main.go:42",
		"msg": "服务",
		"ip": "192.168.1.1",
		"username": "4"
	}

Plain 编码:
默认flags下的示例：
2025/05/16 15:05:34 [INFO] 日志消息
*/
func (l *LogsLogger) outputLog(level LogLevel, skip int, v ...interface{}) {
	var msg string
	item := l.itemPool.Get().(*logItem)
	item.logger = l
	item.level = level
	item.skip = skip + 1

	if _, ok := l.encoder.(StructuredEncoder); ok {
		// JSON 编码
		msg = parseAndEncode(l, skip+1, v...)

		item.msg = msg
		if l.logWriteStrategy == LoggingSync { // 同步写入
			l.output.Write([]byte(msg))
			l.itemPool.Put(item) // 放回对象池
		} else { // 异步写入
			if atomic.LoadInt32(&l.closed) == 1 { // 检查是否已关闭
				l.itemPool.Put(item) // 放回对象池
				return
			}
			select {
			case l.logChan <- item:
			default:
				fmt.Printf("日志通道已满，同步写入该日志: %s", msg)
				l.output.Write([]byte(msg))
				l.itemPool.Put(item) // 放回对象池
			}
		}

	} else {
		// Plain 编码，fields为空，直接输出msg
		msg = l.encoder.Encode(v...)
		if l.hasRootFilePrefix {
			msg = GetRootfilePrefix(skip) + msg
		}

		item.msg = msg
		internalLogger := getLoggerByLevel(l, level)
		if l.logWriteStrategy == LoggingSync { // 同步写入
			internalLogger.Output(skip, msg)
			l.itemPool.Put(item) // 放回对象池
		} else { // 异步写入
			if atomic.LoadInt32(&l.closed) == 1 { // 检查是否已关闭
				l.itemPool.Put(item) // 放回对象池
				return
			}
			item.msg = string(*internalLogger.OutputBuffer(skip, msg))
			select {
			case l.logChan <- item:
			default:
				fmt.Printf("日志通道已满，同步写入该日志: %s", msg)
				internalLogger.Output(skip, msg)
				l.itemPool.Put(item) // 放回对象池
			}
		}
	}
}

func (l *LogsLogger) outputLogw(level LogLevel, skip int, msg string, fields ...Field) {
	item := l.itemPool.Get().(*logItem)
	item.logger = l
	item.level = level
	item.skip = skip + 1

	if _, ok := l.encoder.(StructuredEncoder); ok {
		// JSON 编码
		fullMsg := parseAndEncodeWithFields(l, skip+1, msg, fields...)

		item.msg = fullMsg
		if l.logWriteStrategy == LoggingSync { // 同步写入
			l.output.Write([]byte(fullMsg))
			l.itemPool.Put(item) // 放回对象池
		} else { // 异步写入
			if atomic.LoadInt32(&l.closed) == 1 { // 检查是否已关闭
				l.itemPool.Put(item) // 放回对象池
				return
			}
			select {
			case l.logChan <- item:
			default:
				fmt.Printf("日志通道已满，同步写入该日志: %s", fullMsg)
				l.output.Write([]byte(fullMsg))
				l.itemPool.Put(item) // 放回对象池
			}
		}

	} else {
		// Plain 编码
		if msg != "" {
			fields = append(fields, Field{Key: "msg", Value: msg})
		}
		fullMsg := l.encoder.Encode(fields)
		if l.hasRootFilePrefix {
			fullMsg = GetRootfilePrefix(skip) + fullMsg
		}

		item.msg = fullMsg
		internalLogger := getLoggerByLevel(l, level)
		if l.logWriteStrategy == LoggingSync { // 同步写入
			internalLogger.Output(skip, fullMsg)
			l.itemPool.Put(item) // 放回对象池
		} else { // 异步写入
			if atomic.LoadInt32(&l.closed) == 1 { // 检查是否已关闭
				l.itemPool.Put(item) // 放回对象池
				return
			}
			item.msg = string(*internalLogger.OutputBuffer(skip, fullMsg))
			select {
			case l.logChan <- item:
			default:
				fmt.Printf("日志通道已满，同步写入该日志: %s", fullMsg)
				internalLogger.Output(skip, fullMsg)
				l.itemPool.Put(item) // 放回对象池
			}
		}
	}
}

func (l *LogsLogger) Debug(v ...interface{}) {
	if l.logConf.Level > int(LogLevelDebug) {
		return
	}
	l.outputLog(LogLevelDebug, 3, v)
}

// 格式化日志
func (l *LogsLogger) Debugf(format string, v ...interface{}) {
	if l.logConf.Level > int(LogLevelDebug) {
		return
	}
	l.outputLogf(LogLevelDebug, 3, format, v...)
}

// 结构化日志，含msg
func (l *LogsLogger) Debugw(msg string, fields ...Field) {
	if l.logConf.Level > int(LogLevelDebug) {
		return
	}
	l.outputLogw(LogLevelDebug, 3, msg, fields...)
}

// 结构化日志，不含msg
func (l *LogsLogger) DebugwNoMsg(fields ...Field) {
	if l.logConf.Level > int(LogLevelDebug) {
		return
	}
	l.outputLogw(LogLevelDebug, 3, "", fields...)
}

func (l *LogsLogger) Info(v ...interface{}) {
	if l.logConf.Level > int(LogLevelInfo) {
		return
	}
	l.outputLog(LogLevelInfo, 3, v...)
}

func (l *LogsLogger) Infof(format string, v ...interface{}) {
	if l.logConf.Level > int(LogLevelInfo) {
		return
	}
	l.outputLogf(LogLevelInfo, 3, format, v...)
}

func (l *LogsLogger) Infow(msg string, fields ...Field) {
	if l.logConf.Level > int(LogLevelInfo) {
		return
	}
	l.outputLogw(LogLevelInfo, 3, msg, fields...)
}

func (l *LogsLogger) InfowNoMsg(fields ...Field) {
	if l.logConf.Level > int(LogLevelInfo) {
		return
	}
	l.outputLogw(LogLevelInfo, 3, "", fields...)
}

func (l *LogsLogger) Warn(v ...interface{}) {
	if l.logConf.Level > int(LogLevelWarn) {
		return
	}
	l.outputLog(LogLevelWarn, 3, v...)
}

func (l *LogsLogger) Warnf(format string, v ...interface{}) {
	if l.logConf.Level > int(LogLevelWarn) {
		return
	}
	l.outputLogf(LogLevelWarn, 3, format, v...)
}

func (l *LogsLogger) Warnw(msg string, fields ...Field) {
	if l.logConf.Level > int(LogLevelWarn) {
		return
	}
	l.outputLogw(LogLevelWarn, 3, msg, fields...)
}

func (l *LogsLogger) WarnwNoMsg(fields ...Field) {
	if l.logConf.Level > int(LogLevelWarn) {
		return
	}
	l.outputLogw(LogLevelWarn, 3, "", fields...)
}

func (l *LogsLogger) Error(v ...interface{}) {
	if l.logConf.Level > int(LogLevelError) {
		return
	}
	l.outputLog(LogLevelError, 3, v...)
}
func (l *LogsLogger) Errorf(format string, v ...interface{}) {
	if l.logConf.Level > int(LogLevelError) {
		return
	}
	l.outputLogf(LogLevelError, 3, format, v...)
}

func (l *LogsLogger) Errorw(msg string, fields ...Field) {
	if l.logConf.Level > int(LogLevelError) {
		return
	}
	l.outputLogw(LogLevelError, 3, msg, fields...)
}

func (l *LogsLogger) ErrorwNoMsg(fields ...Field) {
	if l.logConf.Level > int(LogLevelError) {
		return
	}
	l.outputLogw(LogLevelError, 3, "", fields...)
}

func (l *LogsLogger) Fatal(v ...interface{}) {
	l.outputLog(LogLevelFatal, 3, v...)
	os.Exit(1)
}

func (l *LogsLogger) Fatalf(format string, v ...interface{}) {
	l.outputLogf(LogLevelFatal, 3, format, v...)
	os.Exit(1)
}

func (l *LogsLogger) Fatalw(msg string, fields ...Field) {
	l.outputLogw(LogLevelFatal, 3, msg, fields...)
}

func (l *LogsLogger) FatalwNoMsg(fields ...Field) {
	l.outputLogw(LogLevelFatal, 3, "", fields...)
}

func (l *LogsLogger) Panic(v ...interface{}) {
	l.outputLog(LogLevelPanic, 3, v...)
	panic(fmt.Sprint(v...))
}

func (l *LogsLogger) Panicf(format string, v ...interface{}) {
	l.outputLogf(LogLevelPanic, 3, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *LogsLogger) Panicw(msg string, fields ...Field) {
	l.outputLogw(LogLevelPanic, 3, msg, fields...)
}

func (l *LogsLogger) PanicwNoMsg(fields ...Field) {
	l.outputLogw(LogLevelPanic, 3, "", fields...)
}

// 全局日志函数

func Debug(v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelDebug) {
		return
	}
	globalLogger.outputLog(LogLevelDebug, 3, v)
}

// 格式化日志
func Debugf(format string, v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelDebug) {
		return
	}
	globalLogger.outputLogf(LogLevelDebug, 3, format, v...)
}

// 结构化日志，含msg
func Debugw(msg string, fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelDebug) {
		return
	}
	globalLogger.outputLogw(LogLevelDebug, 3, msg, fields...)
}

// 结构化日志，不含msg
func DebugwNoMsg(fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelDebug) {
		return
	}
	globalLogger.outputLogw(LogLevelDebug, 3, "", fields...)
}

func Info(v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelInfo) {
		return
	}
	globalLogger.outputLog(LogLevelInfo, 3, v...)
}

func Infof(format string, v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelInfo) {
		return
	}
	globalLogger.outputLogf(LogLevelInfo, 3, format, v...)
}

func Infow(msg string, fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelInfo) {
		return
	}
	globalLogger.outputLogw(LogLevelInfo, 3, msg, fields...)
}

func InfowNoMsg(fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelInfo) {
		return
	}
	globalLogger.outputLogw(LogLevelInfo, 3, "", fields...)
}

func Warn(v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelWarn) {
		return
	}
	globalLogger.outputLog(LogLevelWarn, 3, v...)
}

func Warnf(format string, v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelWarn) {
		return
	}
	globalLogger.outputLogf(LogLevelWarn, 3, format, v...)
}

func Warnw(msg string, fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelWarn) {
		return
	}
	globalLogger.outputLogw(LogLevelWarn, 3, msg, fields...)
}

func WarnwNoMsg(fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelWarn) {
		return
	}
	globalLogger.outputLogw(LogLevelWarn, 3, "", fields...)
}

func Error(v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelError) {
		return
	}
	globalLogger.outputLog(LogLevelError, 3, v...)
}
func Errorf(format string, v ...interface{}) {
	if globalLogger.logConf.Level > int(LogLevelError) {
		return
	}
	globalLogger.outputLogf(LogLevelError, 3, format, v...)
}

func Errorw(msg string, fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelError) {
		return
	}
	globalLogger.outputLogw(LogLevelError, 3, msg, fields...)
}

func ErrorwNoMsg(fields ...Field) {
	if globalLogger.logConf.Level > int(LogLevelError) {
		return
	}
	globalLogger.outputLogw(LogLevelError, 3, "", fields...)
}

func Fatal(v ...interface{}) {
	globalLogger.outputLog(LogLevelFatal, 3, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	globalLogger.outputLogf(LogLevelFatal, 3, format, v...)
	os.Exit(1)
}

func Fatalw(msg string, fields ...Field) {
	globalLogger.outputLogw(LogLevelFatal, 3, msg, fields...)
}

func FatalwNoMsg(fields ...Field) {
	globalLogger.outputLogw(LogLevelFatal, 3, "", fields...)
}

func Panic(v ...interface{}) {
	globalLogger.outputLog(LogLevelPanic, 3, v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	globalLogger.outputLogf(LogLevelPanic, 3, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func Panicw(msg string, fields ...Field) {
	globalLogger.outputLogw(LogLevelPanic, 3, msg, fields...)
}

func PanicwNoMsg(fields ...Field) {
	globalLogger.outputLogw(LogLevelPanic, 3, "", fields...)
}
