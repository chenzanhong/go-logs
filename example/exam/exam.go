package main

import (

	// "io"
	"os"

	logs "github.com/chenzanhong/logs"
)

func main() {
	// 初始化日志器
	logs.SetupDefault() // 不必要（已经通过init()函数初始化），默认初始化，输出到os.Stdout

	logs.Info("1")
	var name = "World"
	logs.Infof("Hello %s", name)

	// 结构化日志
	logs.SetEncoding(logs.LogEncodingJSON) // 设置日志编码为Json格式，默认是Plain
	logs.Info("message", "key1", "value1", "key2", "value2") // 奇数个参数，第一个参数为消息，后面的参数两两为键值对
	logs.Info("key1", "value1", "key2", "value2")            // 偶数个参数，参数两两为键值对
	// 或者利用提供的结构化对象Field
	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})
	logs.Info("message", field1, field2, field3, field4) 
	// 输出：{"level":"INFO", "timestamp":"2025-05-16 20:42:47.049", "caller":"example/exam/exam.go 28: ", "msg":"message", "key1":"value1", "key2":"value2", "key3":123, "key4":{"key":"value"}}  
	
	// 输出到文件
	file, _ := os.OpenFile("./logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) //path/to/your_logs.log
	logs.SetOutput(file) // 设置日志输出目标

	logs.SetLogLevel(logs.LogLevelInfo) // 设置日志级别为INFO（默认值），输出INFO及以上级别的日志，DEBUG不会输出
	logs.Debug("debug")                 // DEBUG级别日志不会输出

	logs.SetPrefixWithoutDefaultPrefix("[prefix]") // 设置自定义前缀，不包含默认前缀，默认前缀为[INFO]、[WARN]、[ERROR]、[FATAL]、[PANIC]
	logs.Info("info")                              // 输出不含默认的前缀 [INFO]
	logs.SetPrefix("[prefix]")                     // 设置自定义前缀，包含默认前缀
	logs.SetWarnPrefix("这是Warn前缀")                 // 专门设置WARN级别日志的前缀，包含默认前缀

	// 所有日志标志：LogFlagsCommon | Ldate | Ltime | Lmicroseconds | Llongfile | Lshortfile | LUTC | Lmsgprefix | Lrootfile
	logs.SetFlags(logs.LogFlagsCommon | logs.Lrootfile) // 设置日志格式为默认格式（Lmsgprefix | Ldate | Ltim），并包含自定义的相对路径前缀（默认不包含）
	logs.Info("包含了相对路径")                                // 输出包含默认的前缀 [INFO]，并包含自定义的相对路径前缀

	newLogConf := logs.LogConf{ // 创建一个新的日志配置
		Mode:       "console",              // 默认输出到控制台,
		Level:      int(logs.LogLevelInfo), // 默认日志级别为 INFO
		Encoding:   "plain",                // 默认编码为 plain text
		Path:       "",                     // 控制台模式下不需要路径
		MaxSize:    1,                      // 默认每个日志文件最大 10MB，这里设置为 1MB
		MaxBackups: 3,                      // 默认最多保留 3 个备份
		KeepDays:   1,                      // 默认日志文件保留 30 天，这里设置为 1天
		Compress:   false,                  // 默认不压缩旧的日志文件
	}
	newLogger, _ := logs.NewLogger(newLogConf) // 创建一个新的日志器，输出到os.Stdout，日志级别为INFO，日志格式为默认格式（Lmsgprefix | Ldate | Ltim）
	newLogger.Info("这是新的日志器")

	newLogger.SetEncoding(logs.LogEncodingJSON) // 设置日志编码器为 JSON 编码器
	newLogger.Info("这是新的日志器，日志编码器为 JSON 编码器")   // 输出包含默认的前缀 [INFO]，并包含自定义的相对路径前缀，日志格式为 JSON 格式
	// 输出：{"level":"INFO", "timestamp":"2006-01-02 15:04:05.000", "caller":"example/exam/exam.go 58: ", "msg":"这是新的日志器，日志编码器为 JSON 编码器"}

	// 异步日志
	newLogger.SetLogWriteStrategy(logs.LoggingAsync) // 设置日志写入策略为异步写入（管道默认大小1000），默认是同步写入
	for i := 0; i < 1000; i++ {
		logs.Info("异步日志", i) // 异步写入日志，不会阻塞主线程
	}
	newLogger.Close() // 关闭管道，并确保所有异步处理的日志被全部处理完成才结束

	// 重新初始化日志器，输出到os.Stdout，日志级别为INFO，日志格式为默认格式（Lmsgprefix | Ldate | Ltim）
	newLogger.Setup(newLogConf)                                    // 输出到os.Stdout，日志级别为INFO，日志格式为默认格式（Lmsgprefix | Ldate | Ltim），日志编码器为 Plain 编码器，日志写入策略为同步写入
	newLogger.Info("这是重新初始化后的自定义日志器，日志编码器为 Plain 编码器，日志写入策略为同步写入") // 不会Panic
}
