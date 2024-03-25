package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持锁")

	//go:embed lua/unlock.lua
	luaUnlock string

	//go:embed lua/refresh.lua
	luaRefresh string

	//go:embed lua/lock.lua
	luaLock string
)

// Client 就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
	g      singleflight.Group
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) SingleflightLock(ctx context.Context, key string,
	expiration time.Duration,
	timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	for {
		flag := false
		resCh := c.g.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, expiration, timeout, retry)
		})
		select {
		case res := <-resCh:
			if flag {
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context, key string,
	expiration time.Duration,
	timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		//在这里重试
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if res == "OK" {
			return &Lock{
				client:     c.client,
				key:        key,
				lockId:     val,
				expiration: expiration,
				unlockChan: make(chan struct{}, 1),
			}, nil
		}

		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}
		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		//代表的是别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client:     c.client,
		key:        key,
		lockId:     val,
		expiration: expiration,
		unlockChan: make(chan struct{}, 1),
	}, nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	lockId     string
	expiration time.Duration
	unlockChan chan struct{}
}

func (l *Lock) UnLock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.lockId, l.expiration.Seconds()).Int64()
	defer func() {
		//close(l.unlockChan)
		select {
		case l.unlockChan <- struct{}{}:
		default:
			//说明没有人调用 AutoRefresh
		}

	}()
	if err == redis.Nil {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

// 续约
func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.lockId, l.expiration.Seconds()).Int64()

	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

// 自动续约的可控性非常差，因此并不鼓励用户使用这个API
// 如果用户想要万无一失地使用这个分布式锁，
// 那么必须要自己手动调用 Refresh，并且小心处理各种error
func (l *Lock) AutoRefresh(intervel time.Duration, timeout time.Duration) error {

	timeoutChan := make(chan struct{}, 1)

	//间隔多久续约一次
	ticker := time.NewTicker(intervel)

	for {
		select {
		case <-ticker.C:
			//刷新的超时时间怎么设置
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			//出现error了怎么办？
			err := l.Refresh(ctx)
			cancel()

			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}

			if err != nil {
				return err
			}

		case <-timeoutChan:

			//刷新的超时时间怎么设置
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			//出现error了怎么办？
			err := l.Refresh(ctx)
			cancel()

			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}

			if err != nil {

				return err
			}

		case <-l.unlockChan:

			return nil

		}

	}
}
