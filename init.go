package logs

import (
	"fmt"
)

func (i *logItem) Reset() {
	i.logger = nil
	i.level = 0
	i.msg = ""
	i.skip = 0
}

func init() {
	// fmt.Println("初始化日志器")
	if err := globalLogger.Setup(defaultLogConf); err != nil {
		fmt.Printf("Failed to initialize logger: %v", err)
	}
	// fmt.Println("初始化日志器完成")
}
