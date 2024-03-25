package ioc

import (
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/internal/service"
	"GoCamp/webook/pkg/logger"

	"go.uber.org/zap"
)

func InitUserHandler(repo repository.UserRepository) service.UserService {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(l))
}
