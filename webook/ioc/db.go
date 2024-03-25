package ioc

import (
	"GoCamp/webook/internal/repository/dao"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := viper.GetString("db.dsn")
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	db, err := gorm.Open(mysql.Open(dsn))
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
