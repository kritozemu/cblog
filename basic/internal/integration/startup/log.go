package startup

import "compus_blog/basic/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
