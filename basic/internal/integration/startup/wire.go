//go:build wireinject

package startup

import (
	"compus_blog/basic/internal/job"
	"compus_blog/basic/internal/repository"
	"compus_blog/basic/internal/repository/dao"
	"compus_blog/basic/internal/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGORMJobDAO)

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}
