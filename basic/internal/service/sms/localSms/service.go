package localSms

import (
	"context"
	"log"
)

type LocalSms struct {
}

func (s *LocalSms) Send(ctx context.Context, tplID string, args []string, numbers ...string) error {
	log.Println("验证码是：", args)
	return nil
}
