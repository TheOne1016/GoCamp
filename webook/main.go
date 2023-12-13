package main

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/repository/cache"
	"GoCamp/webook/internal/repository/dao"
	"GoCamp/webook/internal/service"
	"GoCamp/webook/internal/service/sms/memory"
	"GoCamp/webook/internal/web"
	"GoCamp/webook/internal/web/middleware"
	"GoCamp/webook/pkg/ginx/middlewares/ratelimit"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	db := initDB()
	rdb := initRedis()
	u := initUser(db, rdb)

	server := initWebServer()

	u.RegisterRoutes(server)

	server.Run(":8081")

}

func initWebServer() *gin.Engine {
	server := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		//Addr: config.Config.Redis.Addr,
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	server.Use(cors.New(cors.Config{
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
	}))

	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("Pe9um9NrZbtVbGzmjIaoMXa4WbY00iuy"), []byte("prsqRXvMwEwrndFyKoiNkaKS0Ua3EKNd"))

	server.Use(sessions.Sessions("mysession", store))
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/signup").Build())
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgnorePaths("/users/login").
	//	IgnorePaths("/users/signup").Build())
	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, *smsSvc) //原 service.NewCodeService(codeRepo, *smsSvc)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		//Addr: config.Config.Redis.Addr,
		Addr: "localhost:6379",
	})
	return redisClient
}

func initDB() *gorm.DB {
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		//一旦初始化过程出错，应用就不要启动了
		panic(err)
	}
	err = dao.InitTble(db)
	if err != nil {
		panic(err)
	}
	return db
}
