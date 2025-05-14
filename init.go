package logs

import "fmt"

func init() {
	if err := SetUp(defaultLogConf); err != nil {
		fmt.Printf("Failed to initialize logger: %v", err)
	}

	initWorkerOnce.Do(func() {
		// fmt.Println("go worker start")
		go worker()
	})
}
