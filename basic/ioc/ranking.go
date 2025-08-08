package ioc

import (
	"compus_blog/basic/internal/job"
	"compus_blog/basic/internal/service"
	"compus_blog/basic/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService,
	rlockClient *rlock.Client, l logger.LoggerV1) *job.RankingJob {
	return job.NewRankingJob(svc, rlockClient, l, time.Second*30)
}

func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
