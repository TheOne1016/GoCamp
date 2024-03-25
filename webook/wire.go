// go:build wireinject

package main

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/repository/cache"
	"GoCamp/webook/internal/repository/dao"
	"GoCamp/webook/internal/service"
	"GoCamp/webook/internal/web"
	ijwt "GoCamp/webook/internal/web/jwt"
	"GoCamp/webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer1() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
