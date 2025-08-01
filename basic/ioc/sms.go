package ioc

import (
	"compus_blog/basic/internal/service/sms"
	"compus_blog/basic/internal/service/sms/memory"
	"compus_blog/basic/internal/service/sms/metrics"
	"compus_blog/basic/internal/service/sms/ratelimit"
	"compus_blog/basic/internal/service/sms/retryable"
	"compus_blog/basic/internal/service/sms/tencent"
	"compus_blog/basic/pkg/limiter"
	"github.com/redis/go-redis/v9"
	"time"
)

// InitSMSService 从内存中实现
func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 换内存，还是换别的
	//svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
	//	limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	//return retryable.NewService(svc, 3)
	// 接入监控
	//return metrics.NewPrometheusDecorator(memory.NewService())
	return memory.NewService()
}

// InitSMSServiceV1 腾讯云实现
func InitSMSServiceV1(cmd redis.Cmdable) sms.Service {
	// client需要自行设定credential
	client := tencent.NewSmsClient()
	// init短信服务
	SMSsvc := tencent.NewSmsFromTencentService(client, tencent.SmsSignName, tencent.SmsAppID)
	// 引入限流机制
	RLsvc := ratelimit.NewRatelimitSMSService(SMSsvc,
		limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	// 引入重试机制
	RTsvc := retryable.NewRetryableSMSService(RLsvc, 3)
	// 引入监控机制
	return metrics.NewPrometheusDecorator(RTsvc)
}
