package logs_test // 改为使用_test后缀的包名

import (
	"os"
	"testing"

	log "github.com/chenzanhong/logs/log_origin"

	"github.com/chenzanhong/logs"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func setup() { // 输出到os.Stdout
	logs.SetupDefault()
}

func setup2() { // 输出到文件
	logs.SetupDefault()
	file, _ := os.OpenFile("../logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	logs.SetOutput(file)
}

func tearDown() {
	logs.Close() // 确保所有异步处理的日志被全部处理完成才结束
}

// -----------------------------输出到os.Stdout
// logrus，同步，Plain基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  42564 ns/op   |       481 B/op    |    15 allocs/op |
// 基准测试：直接记录INFO级别的日志
func BenchmarkLogrusInfo(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("This is an info message")
	}
}

// 基准测试：记录带字段的INFO级别的日志
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  43927 ns/op    |        1327 B/op   |      21 allocs/op |
func BenchmarkLogrusInfoWithFields(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	fields := logrus.Fields{
		"key1": "value1",
		"key2": 123,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("This is an info message with fields")
	}
}

// 基准测试：禁用颜色输出并记录INFO级别的日志
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  49332 ns/op   |      521 B/op     |    15 allocs/op |
func BenchmarkLogrusInfoNoColor(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("This is an info message without color")
	}
}

// 基准测试：使用JSON格式记录INFO级别的日志
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |48670 ns/op      |      907 B/op    |     19 allocs/op|
func BenchmarkLogrusInfoJSON(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Formatter = &logrus.JSONFormatter{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("This is an info message in JSON format")
	}
}

// zap，同步，Plain基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  598.3 ns/op    |       2 B/op      |    0 allocs/op |
func BenchmarkZapSyncPlain(b *testing.B) {
	// 创建一个生产环境级别的logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // 在测试结束时确保所有日志都被刷新

	sugar := logger.Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infow("3")
	}
}

// 原生log基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  47406 ns/op    |        0 B/op    |      0 allocs/opp |
func BenchmarkLogNative(b *testing.B) {
	setup()
	defer tearDown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Print("1")
	}
}

// logs，同步，Plain基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |37794 ns/op      |        16 B/op   |       1 allocs/op|
func BenchmarkLogsSyncPlain(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是同步模式
	logs.SetLogWriteStrategy(logs.LoggingSync)
	logs.SetEncoding(logs.LogEncodingPlain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("2")
	}
}

// logs，异步，Plain基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 38440 ns/op     |      164 B/op    |      3 allocs/op |
func BenchmarkLogsAsyncPlain(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是异步模式
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingPlain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("3")
	}
}

// logs，异步，Json基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 60036 ns/op            2402 B/op         31 allocs/op |
func BenchmarkLogsAsyncJson(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是异步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("服务", "username", "4", "ip", "192.168.1.1")
	}
}

// logs，同步，Json基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// |  51384 ns/op    |      2110 B/op   |     28 allocs/op |
func BenchmarkLogsSyncJson(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是同步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingSync)
	logs.SetEncoding(logs.LogEncodingJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", "5", "ip", "192.168.1.1")
	}
}

// logs，同步，Field基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 66082 ns/op    |      2904 B/op    |    41 allocs/op   |
func BenchmarkLogsSyncField(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是同步模式和JSON编码
	logs.SetEncoding(logs.LogEncodingJSON)

	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", field1, field2, field3, field4)
	}
}

// logs，异步，Field基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 58423 ns/op     |       3120 B/op  |    41 allocs/op   |
func BenchmarkLogsAsyncField(b *testing.B) {
	setup()
	defer tearDown()

	// 确保是异步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingJSON)

	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", field1, field2, field3, field4)
	}
}

// -----------------------------输出到文件
// logs，同步，Plain基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 2036 ns/op              16 B/op          1 allocs/op   |
func BenchmarkLogsSyncPlain2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是同步模式
	logs.SetLogWriteStrategy(logs.LoggingSync)
	logs.SetEncoding(logs.LogEncodingPlain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("2")
	}
}

// logs，异步，Plain基准测试（出现了日志通道已满的情况，defaultLogChanSize = 100）
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 841.9 ns/op     |      164 B/op     |     3 allocs/op   |
func BenchmarkLogsAsyncPlain2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是异步模式
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingPlain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("3")
	}
}

// logs，异步，Json基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 4536 ns/op      |     2434 B/op    |     31 allocs/op  |
func BenchmarkLogsAsyncJson2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是异步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("服务", "username", "4", "ip", "192.168.1.1")
	}
}

// logs，同步，Json基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 5915 ns/op      |      2109 B/op   |    28 allocs/op   |
func BenchmarkLogsSyncJson2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是同步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingSync)
	logs.SetEncoding(logs.LogEncodingJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", "5", "ip", "192.168.1.1")
	}
}

// logs，同步，Field基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 7605 ns/op      |    2903 B/op     |    41 allocs/op  |
func BenchmarkLogsSyncField2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是同步模式和JSON编码
	logs.SetEncoding(logs.LogEncodingJSON)

	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", field1, field2, field3, field4)
	}
}

// logs，异步，Field基准测试
// | 每次操作平均时间 | 每次操作分配的内存 | 每次操作的分配次数 |
// | 4699 ns/op     |      3156 B/op    |     41 allocs/op |
func BenchmarkLogsAsyncField2(b *testing.B) {
	setup2()
	defer tearDown()

	// 确保是异步模式和JSON编码
	logs.SetLogWriteStrategy(logs.LoggingAsync)
	logs.SetEncoding(logs.LogEncodingJSON)

	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logs.Info("username", field1, field2, field3, field4)
	}
}

/*
	所有测试结果汇总
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

	结论：
		1. 原生log性能最好，但是不支持异步、Field
		2. zap性能最好，但是不支持异步、Field
		4. logs性能最好，支持异步、Field，但是输出到文件性能最差
*/
