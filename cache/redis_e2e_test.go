package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_e2e_Set(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.111.133:31379",
		Password: "",
		DB:       0,
	})
	c := NewRedisCache(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := c.Set(ctx, "key1", "abc", time.Minute)
	require.NoError(t, err)
	val, err := c.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "abc", val)
}

func TestRedisCache_e2e_SetV1(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.111.133:31379",
		Password: "",
		DB:       0,
	})

	testCases := []struct {
		name string
		//before func()
		after func(t *testing.T)

		key        string
		val        string
		expiration time.Duration

		wantErr error
	}{
		{
			name: "set value",
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Get(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "val1", res)
				_, err = rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewRedisCache(rdb)
			//tc.before()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			err := c.Set(ctx, tc.key, tc.val, tc.expiration)
			require.NoError(t, err)
			tc.after(t)
		})
	}
}
