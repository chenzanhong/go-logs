package logs

import (
	"bytes"
	"time"
)

/*
    worker 函数是日志处理的主要工作函数，它负责从日志通道中读取日志项，并根据日志级别将其写入到相应的日志器中。
    该函数会在一个无限循环中运行，直到接收到关闭信号。
    在循环中，它会使用 select 语句来等待来自日志通道的日志项或定时器的触发。
    如果接收到日志项，它会根据日志级别将其写入到相应的日志器中。
    如果接收到定时器的触发，它会将当前的日志缓冲区写入到日志文件中。
    如果接收到关闭信号，它会将当前的日志缓冲区写入到日志文件中，并退出循环。
*/
func (l *LogsLogger) worker() {
	defer l.wg.Done()

	batchTimer := time.NewTimer(flushInterval)
	defer batchTimer.Stop()

	for {
		select {
		case item, ok := <-l.logChan:
			if !ok {
				l.itemPool.Put(item)
				l.flushBatch()
				return // channel已关闭，退出
			}

			l.batchMutex.Lock()
			l.batchBuffer = append(l.batchBuffer, []byte(item.msg)) // 假设msg是[]byte类型
			l.batchMutex.Unlock()

			l.itemPool.Put(item)

			if len(l.batchBuffer) >= batchSize {
				l.flushBatch()
				// batchTimer.Reset(flushInterval)
			}
		case <-batchTimer.C:
			l.flushBatch()
			// batchTimer.Reset(flushInterval)
		case <-l.shutdownChan:
			l.flushBatch()
			return // 接收到关闭信号，退出循环
		}
	}
}


/*
    flushBatch 函数用于将当前的日志缓冲区写入到日志文件中。
    该函数会在一个互斥锁的保护下执行，以确保在多线程环境下的线程安全。
    在函数中，它会检查当前的日志缓冲区是否为空，如果为空，则直接返回。
    如果不为空，则会将当前的日志缓冲区写入到日志文件中，并清空日志缓冲区。
    最后，它会释放日志缓冲区的内存。
*/
func (l *LogsLogger) flushBatch() {
	l.batchMutex.Lock()
	defer l.batchMutex.Unlock()

	if len(l.batchBuffer) == 0 {
		return
	}

	// 执行实际的日志写入操作
	msgs := l.batchBuffer
	l.batchBuffer = nil // 清空缓冲区

	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	for _, msg := range msgs {
		buf.Write(msg)
	}

	l.output.Write(buf.Bytes())
}