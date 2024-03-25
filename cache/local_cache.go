package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: 键不存在")
	errKeyExpired  = errors.New("cache: 键过期")
)

type item struct {
	val      any
	deadline time.Time
}

type BuildInMapCacheOption func(cache *BuildInMapCache)

type BuildInMapCache struct {
	data      map[string]*item
	mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(key string, val any)
}

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data:      make(map[string]*item, 100), // 这里可以预估一个容量
		close:     make(chan struct{}),
		onEvicted: func(key string, val any) {},
	}

	for _, opt := range opts {
		opt(res)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				i := 0 // 控制遍历数量
				for key, val := range res.data {

					if i > 1000 {
						break
					}

					if !val.deadline.IsZero() && val.deadline.Before(t) {
						res.delete(key)
					}
					i++
				}

				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

func BuildInMapCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(key, val, expiration)
}

func (b *BuildInMapCache) set(key string, val any, expiration time.Duration) error {

	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	res, ok := b.data[key]
	b.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}

	now := time.Now()
	if !res.deadline.IsZero() && res.deadline.Before(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()

		res, ok = b.data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}

		if !res.deadline.IsZero() && res.deadline.Before(now) {
			b.delete(key)
			return nil, fmt.Errorf("%w, key: %s", errKeyExpired, key)
		}
	}
	return res.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(key)
	return nil
}

func (b *BuildInMapCache) delete(key string) {
	itm, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	b.onEvicted(key, itm.val)
}

func (b *BuildInMapCache) Close() error {
	select {
	case b.close <- struct{}{}:
	default:
		return errors.New("重复关闭")
	}
	return nil
}
