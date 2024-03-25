package ratelimit

import (
	"GoCamp/webook/internal/service/sms"
	"GoCamp/webook/pkg/ratelimit"
	"context"
	"fmt"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		limiter: limiter,
		svc:     svc,
	}
}

func (s *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题， %w", err)
	}
	if limited {
		return errLimited
	}

	err = s.svc.Send(ctx, tpl, args, numbers...)

	return err

}
