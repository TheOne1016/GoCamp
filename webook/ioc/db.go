package ioc

import (
	"GoCamp/webook/internal/repository/dao"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
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
