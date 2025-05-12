package main

import "github.com/chenzanhong/go-logs"

func main() {
	logs.SetupDefault()
	logs.Info("ok")
	logs.Debug("ss")
}