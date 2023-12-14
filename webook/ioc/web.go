package ioc

import (
	"GoCamp/webook/internal/web"
	"GoCamp/webook/internal/web/middleware"
	"GoCamp/webook/pkg/ginx/middlewares/ratelimit"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//AllowOrigins: []string{"http://localhost:8081"},
			//AllowMethods:     []string{"PUT", "PATCH"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
			//不加这个，前端是拿不到的
			ExposeHeaders: []string{"x-jwt-token"},
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
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/signup").Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}
