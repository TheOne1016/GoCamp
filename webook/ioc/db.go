package ioc

import (
	"GoCamp/webook/internal/repository/dao"
	"GoCamp/webook/pkg/logger"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	dsn := viper.GetString("db.dsn")
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//缺了一个 writer
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			//慢查询阈值，只有执行时间超过这个阈值，才会使用
			//50ms 100ms
			//SQL查询必然要求命中索引，最好就是走一次磁盘 IO
			//一次磁盘 IO 是不到 10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries: true,
			LogLevel:                  glogger.Info,
		}),
	})
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

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
