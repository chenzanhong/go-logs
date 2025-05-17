非常感谢你提供的完整 `README.md` 初稿和源码内容。根据你给出的代码结构、功能实现以及你想表达的内容，我将为你优化并**生成一个更完善、更专业、更具可读性的 README.md**。

---

以下是基于你当前项目（包含同步/异步日志、结构化编码、多输出方式、动态配置等）的**最终版 README.md**：

---

# 📚 logs —— 高性能可扩展的日志记录库 for Go

[![GoDoc](https://godoc.org/github.com/chenzanhong/logs?status.svg)](https://pkg.go.dev/github.com/chenzanhong/logs)
[![License](https://img.shields.io/github/license/chenzanhong/logs)](https://github.com/chenzanhong/logs/blob/main/LICENSE)

logs 是一个高性能、可扩展的日志记录库，构建于标准库 `log` 和文件切割库 `lumberjack.v2` 之上，支持多种日志级别、多输出目标（控制台/文件）、结构化编码（JSON/plain）、同步与异步写入模式等功能。

它提供了灵活的配置选项，适用于从简单命令行工具到复杂微服务系统的各种场景。

---

## ✅ 特性

- **多级日志系统**：支持 `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `PANIC` 级别
- **多输出方式**：
  - 控制台输出
  - 文件输出（自动切割归档）
  - 同时输出到控制台和文件
- **多种编码格式**：
  - Plain Text（默认）
  - JSON 格式
- **丰富的日志格式控制**：
  - 自定义前缀
  - 时间戳格式
  - 调用者路径显示（相对路径、短文件名、全路径等）
- **异步日志写入**：提升性能，避免阻塞主流程
- **运行时动态配置更新**：无需重启即可更改日志级别、路径、编码等
- **结构化日志支持**：使用 key-value 形式记录日志信息
- **对象池优化**：减少内存分配，提高性能
- **兼容 go.mod 项目结构**：自动识别项目根目录

---

## 🛠️ 安装

```bash
go get github.com/chenzanhong/logs
```

---

## 🧪 快速开始

### 初始化日志系统（使用默认配置）

```go
package main

import (
    "github.com/chenzanhong/logs"
)

func main() {
    /*
    （可选）使用默认配置初始化日志系统
    if err := logs.SetupDefault(); err != nil {
        panic(err)
    }
    也可以不logs.SetupDefault()，直接调用logs.Info等函数
    因为已经通过init函数进行默认的初始化了
    */
    logs.Info("程序启动成功！")
    logs.Warnf("这是一个警告信息: %v", "test warning")
}
```

---

## ⚙️ 配置说明

### 设置日志级别

```go
logs.SetLogLevel(logs.LogLevelInfo) // 只允许 INFO 及以上级别的日志输出
```

### 设置输出目标

#### 专用的设置输出目标函数

```go
logs.SetOutput(io.Writer)
```

- 输出到控制台：

```go
logs.SetOutput(os.Stdout)
```
- 输出到文件（带切割归档）：

```go
file, err := os.OpenFile("log/myapp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    panic(err)
}
logs.SetOutput(file)
```

- 同时输出到控制台和文件：

```go
file, err := os.OpenFile("log/myapp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err!= nil {
    panic(err)
}
logs.SetOutput(io.MultiWriter(os.Stdout, file))
```

#### 通过修改日志器设置输出目标
- 输出到文件（带切割归档）：

```go
logs.Setup(logs.LogConf{
    Mode:       "file",
    Path:       "log/myapp.log",
    MaxSize:    10,   // MB
    MaxBackups: 5,
    KeepDays:   7,
    Compress:   true,
})
```

- 同时输出到控制台和文件：

```go
logs.Setup(logs.LogConf{
    Mode:     "both",
    Path:     "log/myapp.log",
    Encoding: logs.LogEncodingJSON,
})

```

### 设置日志编码格式

```go
logs.SetEncoding(logs.LogEncodingJSON) // 支持 LogEncodingPlain 或 LogEncodingJSON
```

### 设置自定义日志前缀

```go
// 所有日志添加统一前缀（保留默认级别前缀如 "[INFO] "）
logs.SetPrefix("[MyApp] ")

// 所有日志添加统一前缀（不包含默认级别前缀）
logs.SetPrefixWithoutDefaultPrefix("【日志】")

// 单独设置某个级别前缀（保留默认前缀）
logs.SetErrorPrefix("[错误] ")

// 单独设置某个级别前缀（不保留默认前缀）
logs.SetInfoPrefixWithoutDefaultPrefix("【info】")
```

### 设置日志标志位（Flags）

```go
logs.SetFlags(logs.Ldate | logs.Ltime | logs.Lshortfile)
```

| Flag            | 描述                     |
|------------------|--------------------------|
| `Ldate`          | 输出日期（年/月/日）     |
| `Ltime`          | 输出时间（时/分/秒）     |
| `Lmicroseconds`  | 输出微秒级时间           |
| `Llongfile`      | 输出完整文件名+行号      |
| `Lshortfile`     | 输出短文件名+行号        |
| `LUTC`           | 使用 UTC 时间            |
| `Lmsgprefix`     | 前缀在消息之前           |
| `Lrootfile`      | 显示相对于项目根目录的路径 |
| `LstdFlags`      | 默认值：Ldate \| Ltime   |
| `LogFlagsCommon` | 推荐值：Lmsgprefix \| Ldate \| Ltime |

---

## 📦 日志配置结构体 `LogConf`

```go
type LogConf {
	Mode       string `yaml:"mode"`         // 输出模式：console/file/both
	Level      int    `yaml:"level"`        // 日志级别（int 类型）
	Encoding   string `yaml:"encoding"`     // 编码格式：plain/json
	Path       string `yaml:"path"`         // 日志文件路径
	MaxSize    int    `yaml:"max_size"`     // 单个文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"`  // 最大备份数量
	KeepDays   int    `yaml:"keep_days"`    // 保留天数
	Compress   bool   `yaml:"compress"`     // 是否压缩旧日志
}
```

## 日志器结构体 `LogsLogger`
```go
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
	encoder           Encoder          // 编码器
	logConf           LogConf          // 日志配置
	logWriteStrategy  logWriteStrategy // 默认日志模式为同步模式
	mu                sync.Mutex       // 全局互斥锁

	logChan      chan *logItem
	shutdownOnce sync.Once
	shutdownChan chan struct{}
	itemPool     sync.Pool
	wg           sync.WaitGroup

	batchBuffer [][]byte
	bufferPool  sync.Pool
	batchMutex  sync.Mutex
	batchTicker *time.Ticker
}
```
## 日志项结构体 `logItem`
```go
type logItem struct {
	logger *LogsLogger // 日志器
	level  LogLevel    // 日志级别
	msg    string      // 日志消息
	skip   int         // 调用栈深度
}
```
---

## 🧰 API 方法一览

| 方法                          | 描述                             |
|-------------------------------|----------------------------------|
| `Setup(conf LogConf)`         | 初始化日志配置                   |
| `SetupDefault()`              | 使用默认配置初始化日志系统       |
| `SetOutput(writer io.Writer)` | 设置输出位置                     |
| `SetEncoding(encoding string)`| 设置编码格式（plain/json）       |
| `SetLogLevel(level LogLevel)` | 设置最低输出日志级别             |
| `SetFlags(flags int)`         | 设置日志标志位                   |
| `SetMaxSize(size int)`        | 设置单个日志文件最大大小（MB）   |
| `SetMaxAge(days int)`         | 设置日志保留天数                 |
| `SetMaxBackups(count int)`    | 设置最多保留的备份文件数量       |
| `SetLogWriteStrategy(strategy)`| 设置同步或异步写入               |
| `SetPrefix(prefix string)`    | 设置所有日志级别的通用前缀       |
| `SetXXXPrefix()` / `SetXXXPrefixWithoutDefaultPrefix()` | 分别设置各日志级别的前缀 |
| `Debug(args ...interface{})`  | 输出 DEBUG 级别的日志             |
| `Info(args ...interface{})`   | 输出 INFO 级别的日志              |
| `Warn(args ...interface{})`   | 输出 WARN 级别的日志              |
| `Error(args ...interface{})`  | 输出 ERROR 级别的日志             |
| `Fatal(args...interface{})`  | 输出 FATAL 级别的日志，并退出程序 |
| `Panic(args...interface{})`  | 输出 PANIC 级别的日志，并触发 panic |
| `Debugf(format string, args...interface{})` | 格式化输出 DEBUG 级别的日志       |
| `Infof(format string, args...interface{})`  | 格式化输出 INFO 级别的日志        |
| `Warnf(format string, args...interface{})`  | 格式化输出 WARN 级别的日志        |
| `Errorf(format string, args...interface{})` | 格式化输出 ERROR 级别的日志       |
| `Fatalf(format string, args...interface{})` | 格式化输出 FATAL 级别的日志，并退出程序 |
| `Panicf(format string, args...interface{})` | 格式化输出 PANIC 级别的日志，并触发 panic |
| `Close()`                     | 关闭日志系统（异步模式）         |


> 注：上述方法均作用于全局变量 `globalLogger`，你也可以通过 `NewLogger(conf)` 创建多个独立的日志实例。


| 其他辅助方法                   | 描述                             |
|-------------------------------|----------------------------------|
| `GetRootfilePrefix(skip int)` |    获取根目录前缀(rootfile string) 
| `GetRelativePath(skip int)`   |    获取根目录前缀(file string, line int) |
---

## 📁 默认配置

```go
defaultLogConf = LogConf{
    Mode:       "console",
    Level:      int(LogLevelInfo),
    Encoding:   LogEncodingPlain,
    MaxSize:    10,
    MaxBackups: 3,
    KeepDays:   30,
    Path:       "",
    Compress:   false,
}
```

---

## 📝 示例输出

### Plain 模式

```
2025/05/14 20:19:29 [INFO] example\main.go 32: "ok"
2025/05/14 20:19:29 [ERROR] example\main.go 34: "error"
```

### JSON 模式

```json
{"level":"INFO", "timestamp":"2025-05-17T01:16:05+08:00", "caller":"example/exam/exam.go 21: ", "msg":"message", "key1":"value1", "key2":"value2"}
{"level":"INFO", "timestamp":"2025-05-17T01:16:05+08:00", "caller":"example/exam/exam.go 22: ", "key1":"value1", "key2":"value2"}
```

---

## 📌 进阶用法

### 创建独立日志实例

```go
conf := logs.LogConf{
    Mode:     "both",
    Level:    int(logs.LogLevelDebug),
    Encoding: "plain",
    Path:     "logs/app.log",
}

logger, _ := logs.NewLogger(conf)
logger.SetPrefix("【APP】")
logger.Info("这是另一个日志器输出的信息")
```

---

## 📚 简单的性能运行效果

### 测试环境
MateBook GT 14 笔记本电脑
goos: windows
goarch: amd64
cpu: Intel(R) Core(TM) Ultra 5 125H

### 运行 test/log_benchmark_test.go

#### 输出到 `os.Stdout`
| 测试名称 			 			 | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 | 	说明 			|
| BenchmarkLogNative 			| 47406 ns/op     |      0 B/op      |   0 allocs/op     |  原生log   			|
| BenchmarkLogrusInfo 			| 42564 ns/op     |       481 B/op   |   15 allocs/op    |  logrus，同步，Plain  |
| BenchmarkLogrusInfoWithFields | 43927 ns/op     |       1327 B/op  |   21 allocs/op    |  logrus，同步，Plain |
| BenchmarkLogrusInfoNoColor	| 49332 ns/op     |       521 B/op   |   15 allocs/op    |  logrus，同步，Plain |
| BenchmarkLogrusInfoJSON 		| 48670 ns/op     |       907 B/op   |   19 allocs/op    |  logrus，同步，Json  |
| BenchmarkZapSyncPlain 		| 598.3 ns/op     |       2 B/op     |   0 allocs/op     |  zap，同步，Plain   |
| BenchmarkLogsSyncPlain 		| 37794 ns/op     |       16 B/op    |   1 allocs/op     |  logs，同步，Plain   |
| BenchmarkLogsAsyncPlain 		| 38440 ns/op     |       164 B/op   |   3 allocs/op     |  logs，异步，Plain   |
| BenchmarkLogsAsyncJson 		| 60036 ns/op     |       2402 B/op  |   31 allocs/op    |  logs，异步，Json   |
| BenchmarkLogsSyncJson 		| 51384 ns/op     |       2110 B/op  |   28 allocs/op    |  logs，同步，Json   |
| BenchmarkLogsSyncField 		| 66082 ns/op     |       2904 B/op  |   41 allocs/op    |  logs，同步，Field   |
| BenchmarkLogsAsyncField 		| 58423 ns/op     |       3120 B/op  |   41 allocs/op    |  logs，异步，Field   |

#### 输出到文件
| BenchmarkLogsSyncPlain2 | 2036 ns/op      |       16 B/op    |   1 allocs/op     |  logs，同步，Plain   |
| BenchmarkLogsAsyncPlain2 | 841.9 ns/op     |       164 B/op   |   3 allocs/op     |  logs，异步，Plain   |
| BenchmarkLogsAsyncJson2 | 4536 ns/op      |       2434 B/op  |   31 allocs/op    |  logs，异步，Json   |
| BenchmarkLogsSyncJson2 | 5915 ns/op      |       2109 B/op  |   28 allocs/op    |  logs，同步，Json   |
| BenchmarkLogsSyncField2 | 7605 ns/op      |       2903 B/op  |   41 allocs/op    |  logs，同步，Field   |
| BenchmarkLogsAsyncField2 | 4699 ns/op      |       3156 B/op  |   41 allocs/op    |  logs，异步，Field   |

---

## 📎 注意事项

- 如果启用了 `Lrootfile` 标志，请确保项目根目录存在 `go.mod` 文件。
- 异步写入模式下，务必在程序退出前调用 `logs.Close()` 以确保所有日志被正确写出。
- 日志切割依赖 [lumberjack.v2](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)，请确保其版本兼容性。
- 结构化日志需要传入 key-value 对，例如：`logs.Info("key", "value")`
