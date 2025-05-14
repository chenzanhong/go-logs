package logs

import "fmt"

func init() {
	if err := SetUp(defaultLogConf); err != nil {
		fmt.Printf("Failed to initialize logger: %v", err)
	}

	logChan = make(chan logItem, defaultLogChanSize)
	shutdownChan = make(chan struct{})

	initWorkerOnce.Do(func() {
		wg.Add(1)
		go worker()
	})
}

func worker() {
	defer wg.Done()

	for {
		select {
		case item, ok := <- logChan:
			if !ok {
				return // channel已关闭，退出
			}
			internalLogger := getLoggerByLevel(item.logger, item.level)
			internalLogger.Output(item.skip, item.msg)
		case <-shutdownChan:
			return // 接收到关闭信号，退出循环
		}
	}
}
