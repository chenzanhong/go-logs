package main

import (
	"fmt"

	glog "github.com/chenzanhong/go-logs"
)

func main() {
	// 使用全局日志器
	// glog.SetupDefault()
	fmt.Println("使用全局日志器")
	glog.Info("ok")
	glog.Debug("ss")
	glog.Error("error")

	// 创建另一个默认配置的日志器
	fmt.Println("创建另一个默认配置的日志器")
	logger := glog.NewDefaultLogger()
	logger.Info("ok")
	logger.Debug("ss")
	logger.Error("error")

	// 创建一个新的日志器
	fmt.Println("创建一个新的日志器")
	conf := glog.LogConf{Mode: "both", Level: int(glog.LogLevelDebug), Encoding: "json", Path: "logs/access.log", MaxSize: 10, MaxBackups: 10, KeepDays: 10, Compress: true}
	logger2, err := glog.NewLogger(conf)
	if err != nil {
		fmt.Println("err:", err)
	}
	logger2.Info("ok")
	logger2.Debug("ss")
	logger2.Error("error")
}
