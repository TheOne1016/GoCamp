package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type MiddlewareBuilder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewMiddlewareBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc: fn,
	}
}

func (b *MiddlewareBuilder) AllowReqBody() *MiddlewareBuilder {
	b.allowReqBody = true
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		//URL 本身可能很长
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}

		al := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if b.allowReqBody && ctx.Request.Body != nil {
			//body读完就没有了，所以要重新放回去
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

			if len(body) > 1024 {
				body = body[:1024]
			}
			//这其实是一个很消耗内存和CPU的操作
			//因为会引起复制
			al.ReqBody = string(body)
		}

		if b.allowRespBody {
			ctx.Writer = responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			al.Dutation = time.Since(start).String()
			b.loggerFunc(ctx, al)
		}()

		//执行到业务逻辑
		ctx.Next()

		//b.loggerFunc(ctx, al)

	}
}

// 装饰器模式
type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.al.ReqBody = data
	return w.ResponseWriter.WriteString(data)
}

type AccessLog struct {
	//HTTP 请求的方法
	Method string
	//Url 这个请求的URL
	Url      string
	Dutation string
	ReqBody  string
	RespBody string
	Status   int
}
