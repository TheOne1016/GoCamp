package dao

import "gorm.io/gorm"

func InitTble(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
