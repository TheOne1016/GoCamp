package service

import (
	"GoCamp/webook/internal/domain"
	"GoCamp/webook/internal/repository"
	repomocks "GoCamp/webook/internal/repository/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceLogin(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		//输入
		ctx      context.Context
		email    string
		password string

		//输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$JCXnioDcmusDFI2OeZGp5e04r4OyUUI5CmTtmo2kU6t3Rn0YiYZFa",
						Phone:    "18712345678",
						Ctime:    now,
					}, nil)
				return repo
			},

			email:    "123@qq.com",
			password: "123456@789",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$JCXnioDcmusDFI2OeZGp5e04r4OyUUI5CmTtmo2kU6t3Rn0YiYZFa",
				Phone:    "18712345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},

			email:    "123@qq.com",
			password: "123456@789",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB 错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock DB 错误"))
				return repo
			},

			email:    "123@qq.com",
			password: "123456@789",

			wantUser: domain.User{},
			wantErr:  errors.New("mock DB 错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$JCXnioDcmusDFI2OeZGp5e04r4OyUUI5CmTtmo2kU6t3Rn0YiYZFa",
						Phone:    "18712345678",
						Ctime:    now,
					}, nil)
				return repo
			},

			email:    "123@qq.com",
			password: "000123456@789",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl))
			u, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("123456@789"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
