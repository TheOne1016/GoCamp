package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrCodeSendToMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyToManyTimes = errors.New("验证次数太多")
	ErrUnkonwnForCode        = errors.New("未知错误")
)

// 编译器会在编译的时候，把set_code的代码放进来这个 luaSetCode 变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type RedisCodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
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
		zap.L().Warn("短信发送太频繁",
			zap.String("biz", biz),
			//真正的生产环境中， phone是不能直接记的
			zap.String("phone", phone))
		//你要在对应的告警系统里面配置，
		//比如说规则，一分钟内出现超过100次 WARN，你就告警
		return ErrCodeSendToMany
	default:
		//系统错误
		return errors.New("系统错误")
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		//正常来说，如果频繁出现这个错误，你就要告警，因为有人搞你
		return false, ErrCodeVerifyToManyTimes
	case -2:
		return false, nil
	}
	return false, ErrUnkonwnForCode

}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// LocalCodeCache(homework:implement lua's algo)
// type LocalCodeCache struct {
// 	client redis.Cmdable
// }
