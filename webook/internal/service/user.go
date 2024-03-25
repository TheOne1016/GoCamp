package service

import (
	"GoCamp/webook/internal/domain"
	"GoCamp/webook/internal/repository"
	"GoCamp/webook/pkg/logger"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Edit(ctx context.Context, Id int64, NickName, Birthday, BirefInfo string) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
	GetProfile(ctx context.Context, Id int64) (domain.User, error)
}

type UserServiceOne struct {
	repo repository.UserRepository
	l    logger.LoggerV1
}

func NewUserService(repo repository.UserRepository, l logger.LoggerV1) UserService {
	return &UserServiceOne{repo: repo,
		l: l,
	}
}

func (svc *UserServiceOne) SignUp(ctx context.Context, u domain.User) error {
	//你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//然后就是，存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserServiceOne) Login(ctx context.Context, email, password string) (domain.User, error) {
	//先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	//比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserServiceOne) Edit(ctx context.Context, Id int64, NickName, Birthday, BirefInfo string) error {
	u, err := svc.repo.FindById(ctx, Id)
	if err == repository.ErrUserNotFound {
		return ErrInvalidUserOrPassword
	}
	if err != nil {
		return err
	}
	u.NickName = NickName
	u.Birthday = Birthday
	u.BirefInfo = BirefInfo
	return svc.repo.SaveEditInfo(ctx, u)

}

func (svc *UserServiceOne) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	//这个叫做快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	//要判断，有没有这个用户
	if err != repository.ErrUserNotFound {
		//绝大部分请求进来这里
		//nil 会进来这里
		//不为 ErrUserNotFound 的也会进来这里
		return u, err
	}
	//这里，把 phone 脱敏之后打出来
	//zap.L().Info("用户未注册", zap.String("phone", phone))
	svc.l.Info("用户未注册", logger.String("phone", phone))
	//这个叫做慢路径
	//你明确知道，没有这个用户
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && err != ErrUserDuplicate {
		return u, err
	}
	//因为这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}
func (svc *UserServiceOne) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	//这个叫做快路径
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	//要判断，有没有这个用户
	if err != repository.ErrUserNotFound {
		//绝大部分请求进来这里
		//nil 会进来这里
		//不为 ErrUserNotFound 的也会进来这里
		return u, err
	}
	//这个叫做慢路径
	//你明确知道，没有这个用户
	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && err != ErrUserDuplicate {
		return u, err
	}
	//因为这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *UserServiceOne) GetProfile(ctx context.Context, Id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, Id)
}
