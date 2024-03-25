package web

import (
	"GoCamp/webook/internal/domain"
	"GoCamp/webook/internal/service"
	svcmocks "GoCamp/webook/internal/service/mocks"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService

		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "12345678@",
				}).Return(nil)
				return userSvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"12345678@",
				"confirmPassword":"12345678@"
			}`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			//用不上codeSvc
			h := NewUserHandler(tc.mock(ctrl), nil, nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))

			//assume that no error
			require.NoError(t, err)
			//数据是json格式
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			t.Log(resp)

			//这就是HTTP请求进取GIN框架的入口。
			//当你这样调用的时候，GIN就会处理这个请求
			//响应写回到resp里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}

}
