package service

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

const codeTplId = "1877556"

var (
	ErrCodeSendToMany        = repository.ErrCodeSendToMany
	ErrCodeVerifyToManyTimes = repository.ErrCodeVerifyToManyTimes
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type CodeServiceOne struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &CodeServiceOne{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send发送验证码，我需要什么参数                   //区别业务场景
func (svc *CodeServiceOne) Send(ctx context.Context, biz string, phone string) error {
	//生成一个验证码
	code := svc.generateaCode()
	//塞进去Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		//有问题
		return err
	}

	//发送出去
	svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	// if err != nil{
	// 	//这意味着，Redis 有这个验证码，但是你没发成功，用户根本收不到

	// }

	return nil
}

func (svc *CodeServiceOne) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeServiceOne) generateaCode() string {
	//六位数， num在 0，999999之间，包含0和999999
	num := rand.Intn(1000000)
	//不够六位的，加上前导0
	return fmt.Sprintf("%06d", num)
}
