package main

func main() {

	// db := initDB()
	// rdb := initRedis()
	// u := initUser(db, rdb)

	// server := initWebServer()

	// u.RegisterRoutes(server)

	server := InitWebServer()

	server.Run(":8081")

}

// func initWebServer() *gin.Engine {
// 	server := gin.Default()

// 	redisClient := redis.NewClient(&redis.Options{
// 		//Addr: config.Config.Redis.Addr,
// 		Addr: "localhost:6379",
// 	})
// 	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

// 	server.Use(cors.New(cors.Config{
// 		//AllowOrigins: []string{"http://localhost:8081"},
// 		//AllowMethods:     []string{"PUT", "PATCH"},
// 		AllowHeaders: []string{"Authorization", "Content-Type"},
// 		//不加这个，前端是拿不到的
// 		ExposeHeaders: []string{"x-jwt-token"},
// 		//是否允许你带 cookies之类的东西
// 		AllowCredentials: true,
// 		AllowAllOrigins:  true,
// 		//AllowOriginFunc: func(origin string) bool {
// 		//	if strings.HasPrefix(origin, "http://localhost") {
// 		//		//你的开发环境
// 		//		return true
// 		//	}
// 		//	if strings.HasPrefix(origin, "http://192.168.111") {
// 		//		//你的开发环境
// 		//		return true
// 		//	}
// 		//
// 		//	return strings.Contains(origin, "yourcompany.com")
// 		//},
// 		MaxAge: 12 * time.Hour,
// 	}))

// 	//store := cookie.NewStore([]byte("secret"))
// 	//store := memstore.NewStore([]byte("Pe9um9NrZbtVbGzmjIaoMXa4WbY00iuy"), []byte("prsqRXvMwEwrndFyKoiNkaKS0Ua3EKNd"))

// 	//server.Use(sessions.Sessions("mysession", store))
// 	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
// 		IgnorePaths("/users/login").
// 		IgnorePaths("/users/login_sms/code/send").
// 		IgnorePaths("/users/login_sms").
// 		IgnorePaths("/users/signup").Build())
// 	//server.Use(middleware.NewLoginMiddlewareBuilder().
// 	//	IgnorePaths("/users/login").
// 	//	IgnorePaths("/users/signup").Build())
// 	return server
// }

// func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
// 	ud := dao.NewUserDAO(db)
// 	uc := cache.NewUserCache(rdb)
// 	repo := repository.NewUserRepository(ud, uc)
// 	svc := service.NewUserService(repo)
// 	codeCache := cache.NewCodeCache(rdb)
// 	codeRepo := repository.NewCodeRepository(codeCache)
// 	smsSvc := memory.NewService()
// 	codeSvc := service.NewCodeService(codeRepo, *smsSvc) //原 service.NewCodeService(codeRepo, *smsSvc)
// 	u := web.NewUserHandler(svc, codeSvc)
// 	return u
// }
