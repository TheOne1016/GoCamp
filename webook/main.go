package main

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/repository/cache"
	"GoCamp/webook/internal/repository/dao"
	"GoCamp/webook/internal/service"
	"GoCamp/webook/internal/web"
	"GoCamp/webook/internal/web/jwt"
	"GoCamp/webook/ioc"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {

	// db := initDB()
	// rdb := initRedis()
	// u := initUser(db, rdb)

	// server := initWebServer()

	// u.RegisterRoutes(server)

	InitViperV1()

	initLogger()

	server := InitWebServer()

	server.Run(":8081")

}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	//如果你不 replace, 直接用 zap.L(), 你啥都打不出来
	zap.ReplaceGlobals(logger)
	zap.L().Info("hello, 你搞好了")
}

func InitViperRemote() {
	/*
		这里记得提前再 etcd 里面配置好yaml文件
		etcdctl --endpoints=192.168.111.133:12379 put /webook "$(<dev.yaml)"
	*/
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3",
		//通过 webook 和其它使用etcd的区别出来
		"192.168.111.133:12379", "/webook")
	if err != nil {
		panic(err)
	}

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func InitViperV1() {
	//直接指定文件路径
	viper.SetConfigFile("config/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func InitViper() {
	//配置文件的名字，但是不包含文件扩展名
	//即不包含 .go .yaml 之类的后缀
	viper.SetConfigName("dev")
	// 告诉 viper 我的配置用的是 yaml 格式
	viper.SetConfigType("yaml")
	//当前工作目录下的 config 子目录
	viper.AddConfigPath("./webook/config")
	//读取配置到 viper 里面，或者可以理解为加载到内存里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	//也可以自己重新创建个viper的实例
	// otherViper := viper.New()
	// otherViper.SetConfigName("myjson")
	// otherViper.SetConfigType("json")
	// otherViper.AddConfigPath("./config")

}

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	loggerV1 := ioc.InitLogger()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitMiddlewares(cmdable, loggerV1, handler)
	db := ioc.InitDB()
	userDao := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitWechatService(loggerV1)
	wechatHandlerConfig := ioc.NewWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler, wechatHandlerConfig)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler)
	return engine
}
