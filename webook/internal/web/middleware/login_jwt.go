package middleware

import (
	"GoCamp/webook/internal/web"
	"encoding/gob"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// LoginJWTMiddlewareBuilder JWT登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	//用Go的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		//不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			//没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := segs[1]
		claims := &web.UserClaims{}
		//ParseWithClaims里面一定要传指针
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("Pe9um9NrZbtVbGzmjIaoMXa4WbY00iuy"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		//每10秒钟刷新一次
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("Pe9um9NrZbtVbGzmjIaoMXa4WbY00iuy"))
			if err != nil {
				log.Println("jwt 续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}

		ctx.Set("claims", claims)
	}
}
