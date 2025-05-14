package main

import (
	"fmt"
	// "time"

	glog "github.com/chenzanhong/logs"
	"github.com/chenzanhong/logs/example/exam1"
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
	conf := glog.LogConf{Mode: "both", Level: int(glog.LogLevelDebug), Encoding: "json", Path: "logs/logs.log", MaxSize: 10, MaxBackups: 10, KeepDays: 10, Compress: true}
	logger2, err := glog.NewLogger(conf)
	if err != nil {
		fmt.Println("err:", err)
	}
	a := "1号"
	b := "2号"
	logger2.Infof("这是 %s 的 %s。", a, b)
	logger2.Debug("ss")
	logger2.Error("error")

	fmt.Println("下面是exam输出：")
	exam1.LogExam()

	fmt.Println("回到main函数")
	glog.Info("0")
	glog.Info("1")
	glog.Debug("2")
	glog.Error("3")
	glog.Info("4")
	glog.Debug("5")
	glog.Error("6")
	glog.Info("7")
	glog.Debug("8")
	glog.Error("9")
	glog.Info("0")
	glog.Info("1")
	glog.Debug("2")
	glog.Error("3")
	glog.Info("4")
	glog.Debug("5")
	glog.Error("6")
	glog.Info("7")
	glog.Debug("8")
	glog.Error("9")
	glog.Info("0")
	glog.Info("1")
	glog.Debug("2")
	glog.Error("3")
	glog.Info("4")
	glog.Debug("5")
	glog.Error("6")
	glog.Info("7")
	glog.Debug("8")
	glog.Error("9")
	glog.Close()
}
