package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("邮箱或手机号码冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDao interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, Id int64) (User, error)
	Insert(ctx context.Context, u User) error
	Update(ctx context.Context, u User) error
	FindByWechat(ctx context.Context, openID string) (User, error)
}

type GORMUserDao struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDao {
	return &GORMUserDao{db: db}
}

func (dao *GORMUserDao) FindByWechat(ctx context.Context, openID string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openID).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) FindById(ctx context.Context, Id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", Id).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) Insert(ctx context.Context, u User) error {
	//存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			//邮箱冲突 or 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GORMUserDao) Update(ctx context.Context, u User) error {
	//存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Updates(u).Error
	if _, ok := err.(*mysql.MySQLError); ok {
		//更新失败
		return errors.New("更新数据失败")
	}
	return err
}

//User直接对应数据库表结构

type User struct {
	Id        int64          `gorm:"primaryKey, autoIncrement"`
	Email     sql.NullString `gorm:"unique"`
	Password  string
	NickName  string
	Birthday  string
	BirefInfo string
	//唯一索引允许有多个空值，但是不能有多个空字符串""
	Phone sql.NullString `gorm:"unique"`

	WechatUnionID sql.NullString
	WechatOpenID  sql.NullString `gorm:"unique"`

	//创建时间，毫秒数
	Ctime int64

	//更新时间，毫秒数
	Utime int64
}
