package startup

import "compus_blog/basic/pkg/logger"

func InitLog() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
