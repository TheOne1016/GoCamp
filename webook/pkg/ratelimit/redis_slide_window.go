package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaSlideWindow string

// RedisSlidingWindowLimiter Redis上的滑动窗口算法限流实现
type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable

	//窗口大小
	interval time.Duration
	// 阈值
	rate int

	//interval 内允许 rate 个请求，例如1s内允许3000个请求
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
