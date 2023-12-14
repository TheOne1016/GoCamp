// go:build wireinject

package main

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/repository/cache"
	"GoCamp/webook/internal/repository/dao"
	"GoCamp/webook/internal/service"
	"GoCamp/webook/internal/web"
	"GoCamp/webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		ioc.InitSMSService,

		web.NewUserHandler,

		ioc.InitGin,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
