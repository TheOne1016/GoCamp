package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_e2e_Lock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:31379",
	})

	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		key        string
		expiration time.Duration
		timeout    time.Duration
		retry      RetryStrategy

		wantLock *Lock
		wantErr  error
	}{
		{
			name: "locked",
			before: func(t *testing.T) {
				//
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key1").Result()
				require.NoError(t, err)
				require.True(t, timeout >= time.Second*50)
				_, err = rdb.Del(ctx, "lock_key1").Result()
				require.NoError(t, err)
			},
			key:        "lock_key1",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key:        "lock_key1",
				expiration: time.Minute,
			},
		},
		{
			name: "others hold lock",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "lock_key2", "123", time.Minute).Result()
				require.NoError(t, err)
				require.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				res, err := rdb.GetDel(ctx, "lock_key2").Result()
				require.NoError(t, err)
				require.Equal(t, "123", res)
			},
			key:        "lock_key2",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   3,
			},
			wantErr: fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock),
		},
		{
			name: "retry and locked",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "lock_key3", "123", time.Second*3).Result()
				require.NoError(t, err)
				require.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key3").Result()
				require.NoError(t, err)
				require.True(t, timeout >= time.Second*50)
				_, err = rdb.Del(ctx, "lock_key3").Result()
				require.NoError(t, err)
			},
			key:        "lock_key3",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key: "lock_key3",
			},
		},
	}

	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			lock, err := client.Lock(context.Background(), tc.key, tc.expiration, tc.timeout, tc.retry)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.Equal(t, tc.wantLock.expiration, lock.expiration)
			assert.NotEmpty(t, lock.lockId)
			assert.NotNil(t, lock.client)
			tc.after(t)
		})
	}

}

func TestClient_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:31379",
	})

	testCases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
	}{
		{

			name: "key exist",
			before: func(t *testing.T) {
				// 模拟别人持有锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "val1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "val1", res)
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
		{

			name: "locked",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
				//加锁成功意味着你应该设置好了值
				assert.NotEmpty(t, res)
			},
			key: "key2",
			wantLock: &Lock{
				key:        "key2",
				expiration: time.Minute,
			},
		},
	}

	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.before(t)
			lock, err := client.TryLock(ctx, tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.Equal(t, tc.wantLock.expiration, lock.expiration)
			assert.NotEmpty(t, lock.lockId)
			assert.NotNil(t, lock.client)
			tc.after(t)
		})
	}
}

func TestClient_e2e_UnLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:31379",
	})

	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {},
			lock: &Lock{
				key:    "unlock_key1",
				lockId: "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "val2", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				assert.Equal(t, "val2", res)
			},
			lock: &Lock{
				key:    "unlock_key2",
				lockId: "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlock",
			before: func(t *testing.T) {

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &Lock{
				key:    "unlock_key3",
				lockId: "123",
				client: rdb,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.before(t)
			err := tc.lock.UnLock(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestClient_e2e_Refresh(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:31379",
	})

	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {},
			lock: &Lock{
				key:        "refresh_key1",
				lockId:     "123",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "refresh_key2", "val2", time.Second*10).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "refresh_key2").Result()
				require.NoError(t, err)
				//如果要是刷新成功了，过期时间是一分钟，即便考虑测试本身的运行时间，timeout > 10s
				//也就是 ， 如果 timeout < 10s，说明没有刷新成功
				require.True(t, timeout <= time.Second*10)
				_, err = rdb.Del(ctx, "refresh_key2").Result()
				require.NoError(t, err)

			},
			lock: &Lock{
				key:        "refresh_key2",
				lockId:     "123",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "refreshed",
			before: func(t *testing.T) {

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				res, err := rdb.Set(ctx, "refresh_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "refresh_key3").Result()
				require.NoError(t, err)
				//如果要是刷新成功了，过期时间是一分钟，即便考虑测试本身的运行时间，timeout > 10s

				require.True(t, timeout > time.Second*50)
				_, err = rdb.Del(ctx, "refresh_key3").Result()
				require.NoError(t, err)
			},
			lock: &Lock{
				key:        "refresh_key3",
				lockId:     "123",
				client:     rdb,
				expiration: time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.before(t)
			err := tc.lock.Refresh(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}
