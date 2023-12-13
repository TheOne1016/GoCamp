package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendToMany = errors.New("发送验证码太频繁")
)

// 编译器会在编译的时候，把set_code的代码放进来这个 luaSetCode 变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		//毫无问题
		return nil
	case -1:
		// 发送太频繁
		return ErrCodeSendToMany
	default:
		//系统错误
		return errors.New("系统错误")
	}
}

func (c *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {

}

func (c *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
