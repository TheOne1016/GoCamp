package domain

import "time"

//领域对象

type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	NickName   string
	Birthday   string
	BirefInfo  string
	WechatInfo WechatInfo
	Ctime      time.Time
}
