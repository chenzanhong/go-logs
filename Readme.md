logs/
├── logger.go            // 主入口
├── encoder.go           // PlainEncoder / JsonEncoder
├── flags.go             // SetFlags 相关
├── config.go            // LogConf / SetUp / DefaultLogConf
├── output.go            // outputLog / outputLogf
├── path.go              // GetRelativePath / findProjectRoot
└── levels.go            // LogLevel / SetLogLevel （新增）