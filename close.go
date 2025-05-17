package logs

import "sync/atomic"

func (l *LogsLogger) Close() {
	l.closeOnce.Do(func() {
		// 标记为已关闭
		atomic.StoreInt32(&l.closed, 1)

		// 关闭 logChan，防止新日志进来
		close(l.logChan)

		// 通知 workers 停止
		l.shutdownOnce.Do(func() {
			close(l.shutdownChan)
		})

		// 强制刷新一次（可选）
		l.flushBatch()

		// 等待所有 worker 退出
		l.wg.Wait()
	})
	
	l.Setup(l.logConf) // 重新设置日志器，确保日志器可以正常工作（可选）
}

func Close() {
	globalLogger.Close()
}