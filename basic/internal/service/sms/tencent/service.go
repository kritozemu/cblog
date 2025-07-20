package tencent

import (
	sms2 "compus_blog/basic/internal/service/sms"
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
)

const (
	SmsSignName = "xx科技"
	SmsAppID    = "100000000"
)

type SmsFromTencentService struct {
	client   *sms.Client
	signName *string
	appID    *string
}

func NewSmsFromTencentService(client *sms.Client, signName string, appID string) sms2.Service {
	return &SmsFromTencentService{
		client:   client,
		signName: ekit.ToPtr[string](signName),
		appID:    ekit.ToPtr[string](appID),
	}
}

func (s *SmsFromTencentService) Send(ctx context.Context, tplID string, args []string, numbers ...string) error {
	request := sms.NewSendSmsRequest()

	request.SetContext(ctx)
	request.SmsSdkAppId = s.appID
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr[string](tplID)
	request.TemplateParamSet = s.toPtrString(args)
	request.TemplateParamSet = s.toPtrString(numbers)

	response, err := s.client.SendSms(request)
	// 处理异常
	if err != nil {
		return err
	}

	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进入到这
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败code:%s,msg:%s", *status.Code, *status.Message)
		}
	}
	return nil
}

// NewSmsClient KEY从环境变量里取
func NewSmsClient() *sms.Client {
	credential := common.NewCredential(
		os.Getenv("TENCENTCLOUD_SECRET_ID"),
		os.Getenv("TENCENTCLOUD_SECRET_KEY"),
	)
	client, err := sms.NewClient(credential, "ap-shenzhen", profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return client
}

func (s *SmsFromTencentService) toPtrString(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return &src
	})
}
