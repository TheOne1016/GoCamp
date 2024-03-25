package ioc

import (
	"GoCamp/webook/internal/web"
	ijwt "GoCamp/webook/internal/web/jwt"
	"GoCamp/webook/internal/web/middleware"
	"GoCamp/webook/pkg/ginx/middlewares/logger"
	logger2 "GoCamp/webook/pkg/logger"
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server) //这一句可以注释掉，微信登陆的
	return server
}

func InitMiddlewares(redisClient redis.Cmdable,
	l logger2.LoggerV1,
	jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//AllowOrigins: []string{"http://localhost:8081"},
			//AllowMethods:     []string{"PUT", "PATCH"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
			//不加这个，前端是拿不到的
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			//是否允许你带 cookies之类的东西
			AllowCredentials: true,
			AllowAllOrigins:  true,
			//AllowOriginFunc: func(origin string) bool {
			//	if strings.HasPrefix(origin, "http://localhost") {
			//		//你的开发环境
			//		return true
			//	}
			//	if strings.HasPrefix(origin, "http://192.168.111") {
			//		//你的开发环境
			//		return true
			//	}
			//
			//	return strings.Contains(origin, "yourcompany.com")
			//},
			MaxAge: 12 * time.Hour,
		}),
		logger.NewMiddlewareBuilder(func(ctx context.Context, al *logger.AccessLog) {
			l.Debug("HTTP 请求", logger2.Field{Key: "al", Value: al})
		}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/signup").Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}
