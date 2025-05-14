package logs

func Close() {
	// 关闭 shutdownChan，通知 worker 停止接收新日志
	shutdownOnce.Do(func() {
		close(shutdownChan)
	})

	// 关闭 logChan, 防止新的日志继续写入
	close(logChan)

	// 等待 worker 处理完剩余日志
	wg.Wait()
}
