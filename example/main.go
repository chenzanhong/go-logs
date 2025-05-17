package main

import (
	// "time"
	"fmt"

	log "github.com/chenzanhong/logs/log_origin"

	"github.com/chenzanhong/logs"
)

// 原生log
func log1() {
	var a = 1
	for i := 0; i < 80000; i++ {
		log.Print("1")
		a++
	}
	fmt.Println("--------------------------------------", a)
}

// logs，Plain
func log2() {
	var b = 1
	for i := 0; i < 80000; i++ {
		logs.Info("2")
		b++
	}
	fmt.Println("--------------------------------------", b)
}

// logs，Json
func log3() {
	var d = 1
	for i := 0; i < 80000; i++ {
		logs.Info("服务", "username", "4", "ip", "192.168.1.1")
		d++
	}
	fmt.Println("--------------------------------------", d)
}

// logs，Field
func log4() {
	var f = 1
	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value"})
	for i := 0; i < 80000; i++ {
		logs.Info("message", field1, field2, field3, field4)
		f++
	}
	fmt.Println("--------------------------------------", f)
}

func main() {

	// Plain编码
	field1 := logs.Field{Key: "key1", Value: "value1"}
	field2 := logs.String("key2", "value2")
	field3 := logs.Int("key3", 123)
	field4 := logs.Any("key4", map[string]interface{}{"key": "value", "key2": "name2"})
	logs.Info("message", field1, field2, field3, field4)
	logs.Infow("message", field1, field2, field3, field4) // 结构化
	logs.InfowNoMsg(field1, field2, field3, field4)       // 结构化，不含msg字段

	fmt.Println()

	// JSON编码
	logs.SetEncoding(logs.LogEncodingJSON)
	logs.Info("message", field1, field2, field3, field4)
	logs.Infow("message", field1, field2, field3, field4) // 结构化
	logs.InfowNoMsg(field1, field2, field3, field4)       // 结构化，不含msg字段

	// 设置输出到文件D:\项目\logs\logs\logs.log
	// file, _ := os.OpenFile("D:\\项目\\logs\\logs\\logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// logs.SetOutput(file)
	// var start1 = time.Now()
	// log1()
	// time1 := time.Since(start1)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 同步、Plain
	// var start2 = time.Now()
	// log2()
	// time2 := time.Since(start2)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 异步、Plain
	// logs.SetLogWriteStrategy(logs.LoggingAsync)
	// var start3 = time.Now()
	// log2()
	// time3 := time.Since(start3)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 异步、Json
	// logs.SetEncoding(logs.LogEncodingJSON)
	// var start4 = time.Now()
	// log3()
	// time4 := time.Since(start4)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 同步、Json
	// logs.SetLogWriteStrategy(logs.LoggingSync)
	// var start5 = time.Now()
	// log3()
	// time5 := time.Since(start5)
	// runtime.GC()
	// time.Sleep(1 * time.Second)
	// logs.Close() // 确保所有异步处理的日志被全部处理完成才结束
	// logs.Infof("你好，%s", "世界")

	// fmt.Println("原生log：", time1)
	// fmt.Println("logs，同步，Plain：", time2)
	// fmt.Println("logs，异步，Plain：", time3)
	// fmt.Println("logs，异步、Json：", time4)
	// fmt.Println("logs，同步、Json：", time5)

	// // 同步、Json、Field
	// var start6 = time.Now()
	// log4()
	// time6 := time.Since(start6)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 异步、Json、Field
	// logs.SetLogWriteStrategy(logs.LoggingAsync)
	// var start7 = time.Now()
	// log4()
	// time7 := time.Since(start7)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 异步、Plain、Field
	// logs.SetEncoding(logs.LogEncodingPlain) // 设置日志编码器为 Plain 编码器
	// var start8 = time.Now()
	// log4()
	// time8 := time.Since(start8)
	// runtime.GC()
	// time.Sleep(1 * time.Second)

	// // 同步、Plain、Field
	// logs.SetLogWriteStrategy(logs.LoggingSync) // 设置日志写入策略为同步写入（管道默认大小1000），默认是同步写入
	// var start9 = time.Now()
	// log4()
	// time9 := time.Since(start9)
	// runtime.GC()

	// fmt.Println("logs，同步、Json、Field：", time6)
	// fmt.Println("logs，异步、Json、Field：", time7)
	// fmt.Println("logs，异步、Plain、Field：", time8)
	// fmt.Println("logs，同步、Plain、Field：", time9)
	// logs.Close() // 确保所有异步处理的日志被全部处理完成才结束
}
