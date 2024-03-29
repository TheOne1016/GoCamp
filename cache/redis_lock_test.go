package cache

import (
	"GoCamp/cache/mocks"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestClient_Lock(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) redis.Cmdable
		key      string
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := NewClient(tc.mock(ctrl))
			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.Equal(t, tc.wantLock.expiration, l.expiration)

			assert.NotEmpty(t, l.lockId)
		})
	}
}

func TestClient_UnLock(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable
		key  string
		val  string

		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).
					Return(res)
				return cmd

			},
			key:     "key1",
			val:     "val1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).
					Return(res)
				return cmd

			},
			key:     "key1",
			val:     "val1",
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).
					Return(res)
				return cmd

			},
			key: "key1",
			val: "val1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				client: tc.mock(ctrl),
				key:    tc.key,
				lockId: tc.val,
			}
			err := lock.UnLock(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestClient_Refresh(t *testing.T) {
	testCases := []struct {
		name string

		mock       func(ctrl *gomock.Controller) redis.Cmdable
		key        string
		val        string
		expiration time.Duration

		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"val1", float64(60)}).
					Return(res)
				return cmd

			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"val1", float64(60)}).
					Return(res)
				return cmd

			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    ErrLockNotHold,
		},
		{
			name: "refreshed",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"val1", float64(60)}).
					Return(res)
				return cmd

			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				client:     tc.mock(ctrl),
				key:        tc.key,
				lockId:     tc.val,
				expiration: tc.expiration,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

// 演示如何使用Refresh方法
func ExampleLock_Refresh() {
	// 加锁成功，你拿到了一个 Lock
	var l *Lock
	stopChan := make(chan struct{})
	errChan := make(chan error)
	timeoutChan := make(chan struct{}, 1)
	go func() {
		//间隔多久续约一次
		ticker := time.NewTicker(time.Second * 10)
		timeoutRetry := 0
		for {
			select {
			case <-ticker.C:
				//刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//出现error了怎么办？
				err := l.Refresh(ctx)
				cancel()

				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}

				if err != nil {
					errChan <- err
					//close(stopChan)
					//close(errChan)
					return
				}
				timeoutRetry = 0

			case <-timeoutChan:
				timeoutRetry++
				if timeoutRetry > 10 {
					errChan <- context.DeadlineExceeded
					return
				}
				//刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//出现error了怎么办？
				err := l.Refresh(ctx)
				cancel()

				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}

				if err != nil {
					errChan <- err
					//close(stopChan)
					//close(errChan)
					return
				}

			case <-stopChan:
				l.UnLock(context.Background())
				return

			}

		}
	}()

	//这边假设是你的业务

	//你执行业务，有很复杂的逻辑, 要记得在中间步骤检测 errChan 有没有信号
	//循环的例子——加入你的业务是循环处理
	for i := 0; i < 100; i++ {
		select {
		//这里，续约失败
		case <-errChan:
			break
		default:
			//正常的业务逻辑
		}
	}

	//如果没有循环，那就是每个步骤内都检测一下
	select {
	case err := <-errChan:
		//续约失败，你要终端业务
		log.Fatal(err)
		return
	default:
		//这是你的步骤1
	}

	select {
	case err := <-errChan:
		//续约失败，你要终端业务
		log.Fatal(err)
		return
	default:
		//这是你的步骤2
	}

	select {
	case err := <-errChan:
		//续约失败，你要终端业务
		log.Fatal(err)
		return
	default:
		//这是你的步骤3。。。
	}

	//你的业务结束了，就要退出续约的循环了
	stopChan <- struct{}{}

	fmt.Println("Hello")
	// Output:
	// Hello
}

func ExampleLock_AutoRefresh() {
	var l *Lock
	go func() {
		// 这里返回 error 了，你要中断业务
		l.AutoRefresh(time.Second*10, time.Second)
	}()
}
