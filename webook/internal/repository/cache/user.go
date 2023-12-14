package cache

import (
	"GoCamp/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	//面向接口
	//传单机Redis可以，传cluster的Redis也可以
	client     redis.Cmdable
	expiration time.Duration
}

var ErrKeyNotExist = redis.Nil

// NewUserCache
// A用到了B，B一定是接口 --> 这个是保证面向接口
// A用到了B，B一定是A的字段 --> 规避包变量、包方法，都非常缺乏扩展性
// A用到了B，A绝对不初始化B，而是外面注入 --> 保持依赖注入(DI, Dependency Injection)和依赖反转(IOC)
// func NewUserCache(client redis.Cmdable, expiration time.Duration) *UserCache {
// 	return &UserCache{
// 		client:     client,
// 		expiration: time.Minute * 15,
// 	}
// }

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// 只要error为nil，就认为缓存里有数据
// 如果没有数据，返回一个特定的error
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}

	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}

	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
