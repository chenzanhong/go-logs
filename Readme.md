# 📚 go-Logs —— 可扩展的日志记录库 for Go

[![GoDoc](https://godoc.org/github.com/chenzanhong/logs?status.svg)](https://godoc.org/github.com/chenzanhong/logs)
[![License](https://img.shields.io/github/license/chenzanhong/logs)](https://github.com/chenzanhong/logs/blob/main/LICENSE)

go-Logs 是一个基于log和lumberjack.v2、可配置的 Go 语言日志记录库，支持多种日志级别、多输出目标（控制台/文件）、结构化编码（JSON/plain）以及同步/异步写入模式等功能。

---

## ✅ 特性

- 支持日志级别：`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `PANIC`
- 多种输出方式：
  - 控制台输出
  - 文件输出（带自动切割归档）
  - 同时输出到控制台和文件
- 支持日志格式：
  - Plain Text（默认）
  - JSON 格式
- 自定义日志前缀（或默认的前缀）、时间戳格式、调用者路径等
- 默认同步写入日志（可切换为异步步）
- 支持运行时动态修改配置（如日志路径、编码、级别等）

---

## 🛠️ 安装

```bash
go get github.com/chenzanhong/logs
```

---

## 🧪 快速使用示例

### 基本初始化

```go
package main

import (
    "github.com/chenzanhong/logs"
)

func main() {
    /*
    使用默认配置初始化日志系统
    也可以不logs.SetupDefault()，直接调用logs.Info等函数
    因为已经通过init函数进行默认的初始了
    */
    err := logs.SetupDefault()
    if err != nil {
        panic(err)
    }

    logs.Info("程序启动成功！")
    logs.Warnf("这是一个警告信息: %v", "test warning")
}
```

---

## ⚙️ 配置说明

### 设置日志级别

```go
logs.SetLogLevel(logs.LogLevelInfo) // 允许 INFO 及以上级别的日志输出
```

### 设置输出方式

- 输出到控制台：

```go
logs.SetOutput(os.Stdout)
```

- 输出到文件：

```go
logs.SetUp(logs.LogConf{
    Mode:     "file",
    Path:     "/var/log/myapp.log",
    MaxSize:  10,  // MB
    MaxBackups: 5,
    KeepDays: 7,
    Compress: true,
})
```

- 同时输出到控制台和文件：

```go
logs.SetUp(logs.LogConf{
    Mode:     "both",
    Path:     "/var/log/myapp.log",
    Encoding: logs.LogEncodingPlain,
})
```

- 或者调用下面的函数直接设置输出对象

```go
logs.SetOutput(io.Writer)
```

### 设置日志编码方式

```go
logs.SetEncoding(logs.LogEncodingJSON) // 或 LogEncodingPlain
```

### 设置自定义前缀

```go
logs.SetPrefix("[MyApp] ") // 所有级别的日志添加统一前缀，并带有默认的前缀（Info类型的默认前缀为"[INFO] "，其他类似）
logs.SetPrefixWithoutDefaultPrefix("[日志]") // 所有级别的日志添加统一前缀，不带默认前缀
logs.SetErrorPrefix("[错误] ") // 单独设置 ERROR 的前缀，带默认前缀
logs.SetInfoPrefixWithoutDefaultPrefix("【info】") // 单独设置 INFO 的前缀，不带默认前缀
```

### 设置日志标志（Flags）

```go
logs.SetFlags(logs.Ldate | logs.Ltime | logs.Lshortfile)
```

可用标志：

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
| `LstdFlags`      | 日期和时间               |
| `LogFlagsCommon` | 默认值：Lmsgprefix | Ldate | Ltime | Lrootfile |

---

## 📦 结构体与接口

```go
type LogConf struct {
	Mode       string `yaml:"mode"`        // 日志输出模式：console/file/both
	Level      int    `yaml:"level"`       // 日志级别：debug/info/warn/error/fatal/panic
	Encoding   string `yaml:"encoding"`    // 日志编码：plain/json
	Path       string `yaml:"path"`        // 日志文件路径（仅在file或both模式下使用）
	MaxSize    int    `yaml:"max_size"`    // 日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 日志文件最大保留数量
	KeepDays   int    `yaml:"keep_days"`   // 日志文件保留天数（仅在file或both模式下使用）
	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件（仅在file或both模式下使用）
}

type LogsLogger struct {
    logConf         LogConf
    encoder         Encoder
    output          io.Writer
    debugL          *log.Logger
    infoL           *log.Logger
    warnL           *log.Logger
    errorL          *log.Logger
    fatalL          *log.Logger
    panicL          *log.Logger
    logFlags        int
    hasRootFilePrefix bool
    logWriteStrategy logWriteStrategy
}
```

---

## 🧰 API 方法一览

| 方法名                          | 描述                             |
|----------------------------------|----------------------------------|
| `SetUp(conf LogConf)`            | 初始化日志配置                   |
| `SetOutput(writer io.Writer)`    | 设置输出位置并自动识别输出模式   |
| `SetEncoding(encoding string)`   | 设置日志编码（plain/json）       |
| `SetLogLevel(level LogLevel)`    | 设置最低输出日志级别             |
| `SetFlags(flags int)`            | 设置日志标志位                   |
| `SetMaxSize(size int)`           | 设置单个日志文件最大大小（MB）   |
| `SetMaxAge(days int)`            | 设置日志保留天数                 |
| `SetMaxBackups(count int)`       | 设置最多保留的备份文件数量       |
| `SetLogWriteStrategy(strategy)`  | 设置同步或异步写入               |
| `SetPrefix(prefix string)`       | 设置所有日志级别的通用前缀       |
| `SetXXXPrefix()` / `SetXXXPrefixWithoutDefaultPrefix()` | 分别设置各日志级别的前缀 |


上述的函数设置都是基于全局的globalLogger logs.LogsLogger
logs.LogsLogger实现了上述的所有方法
可以通过logs.NewLogger(conf LogConf) 创建一个自定义的logs.LogsLogger，并调用上述方法

```go
    conf := logs.LogConf{Mode: "both", Level: int(glog.LogLevelDebug), Encoding: "plain", Path: "logs/logs.log", MaxSize: 10, MaxBackups: 10, KeepDays: 10, Compress: true}
	logger, err := logs.NewLogger(conf)
	if err != nil {
		fmt.Println("err:", err)
	}
    logger.SetPrefix("【日志】")
    logger.Info("执行成功")

```

---

## 📁 默认日志配置

```go
defaultLogConf = LogConf{
    Mode:     "console",
    Level:    int(LogLevelInfo),
    Encoding: LogEncodingPlain,
    MaxSize:  10, // 10MB
    MaxBackups: 3,
    KeepDays: 30,
    Path:     "",
    Compress: false,
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
{
  "level": "info",
  "timestamp": "2025-05-14T22:10:00Z",
  "file": "main.go:12",
  "message": "程序启动成功！"
}
```

---

## 📎 注意事项

- 如果使用 `Lrootfile` 标志，请确保项目根目录存在 `go.mod` 文件。
- 异步写入模式时为确保所有日志在程序结束前被处理，请调用logs.Close()
- 日志文件切割依赖 [lumberjack.v2](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)，请确保其版本兼容性。

---

## 📣 贡献指南

欢迎贡献代码、文档、测试案例或提出 Issue！

- Fork 仓库
- 创建新分支 (`git checkout -b feature/new-feature`)
- 提交更改 (`git commit -am 'Add new feature'`)
- 推送到远程分支 (`git push origin feature/new-feature`)
- 创建 Pull Request

---