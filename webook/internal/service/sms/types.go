package sms

import "context"

// Service 发送短信的抽象
// 目前可以理解为，这是一个为了适配不同的短信供应商的抽象
type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}

type NamedArg struct {
	Val  string
	Name string
}
