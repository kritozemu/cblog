package retryable

import (
	"compus_blog/basic/internal/service/sms"
	"context"
	"errors"
)

type RetryableSMSService struct {
	svc sms.Service
	//重试机制
	duration int
}

func NewRetryableSMSService(svc sms.Service, duration int) sms.Service {
	return &RetryableSMSService{
		svc:      svc,
		duration: duration,
	}
}

func (s *RetryableSMSService) Send(ctx context.Context, tplID string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tplID, args, numbers...)
	cnt := 1
	if err != nil && cnt < s.duration {
		err = s.svc.Send(ctx, tplID, args, numbers...)
		if err != nil {
			cnt++
		}
	}
	return errors.New("重试都失败了")
}
