package failover

import (
	"GoCamp/webook/internal/service/sms"
	"context"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	//你的服务商
	svcs []sms.Service
	idx  int32
	// 连续超时的个数
	cnt int32

	//阈值
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, cnt int32) sms.Service {
	return &TimeoutFailoverSMSService{
		svcs: svcs,
		cnt:  cnt,
	}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	case nil:
		//连续状态被打断
		atomic.StoreInt32(&t.cnt, 0)
	default:
		//不知道什么错误

	}
	return err
}
