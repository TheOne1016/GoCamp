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
)

type BuildInMapCache struct {
	data  map[string]any
	mutex sync.RWMutex
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data[key] = val
	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	res, ok := b.data[key]
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}
	return res, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {

}
