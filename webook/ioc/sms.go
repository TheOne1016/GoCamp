package ioc

import (
	"GoCamp/webook/internal/service/sms"
	"GoCamp/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
