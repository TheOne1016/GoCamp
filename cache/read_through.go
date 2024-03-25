package cache

import (
	"context"
	//"internal/singleflight"
	"log"
	"time"
)

// ReadThroughCache 是一个装饰器
// 在原本Cache的功能上添加了 read through功能
type ReadThroughCache struct {
	Cache
	LoadFunc func(ctx context.Context, key string) (any, error)
	//g        singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			if er := r.Cache.Set(ctx, key, val, time.Minute); er != nil {
				log.Fatalf("刷新缓存失败，err: %v", er)
			}
		}
	}
	return val, err
}

// func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
// 	val, err := r.Cache.Get(ctx, key)
// 	if err == errKeyNotFound {
// 		val, err, _ = r.g.Do(key, func() (interface{}, error) {
// 			v, er := r.LoadFunc(ctx, key)
// 			if er == nil {
// 				if er := r.Cache.Set(ctx, key, val, time.Minute); er != nil {
// 					log.Fatalf("刷新缓存失败，err: %v", er)
// 				}
// 			}
// 			return v, er
// 		})

// 	}
// 	return val, err
// }
