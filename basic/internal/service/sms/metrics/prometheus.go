package metrics

import (
	"compus_blog/basic/internal/service/sms"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) sms.Service {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "kz",
		Subsystem:  "cblog",
		Name:       "sms_resp_time",
		Help:       "统计 SMS 服务的性能数据",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (s *PrometheusDecorator) Send(ctx context.Context,
	tplID string, args []string, numbers ...string) error {
	start := time.Now()

	defer func() {
		duration := time.Since(start).Milliseconds()
		s.vector.WithLabelValues("send").Observe(float64(duration))
	}()

	return s.svc.Send(ctx, tplID, args, numbers...)
}
