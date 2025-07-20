package tencent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSender(t *testing.T) {
	c := NewSmsClient()
	s := NewSmsFromTencentService(c, SmsSignName, SmsAppID)

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:   "发送验证码",
			tplId:  "模板id",
			params: []string{"123456"},
			// 改成你的手机号码
			numbers: []string{"174xxxx7756"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
