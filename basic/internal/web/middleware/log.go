package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddleWareBuilder struct {
	logFn         func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddleWareBuilder(logFn func(ctx context.Context, al AccessLog)) *LogMiddleWareBuilder {
	return &LogMiddleWareBuilder{logFn: logFn}
}

func (l *LogMiddleWareBuilder) AllowReqBody() *LogMiddleWareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddleWareBuilder) AllowRespBody() *LogMiddleWareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddleWareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := c.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			body, _ := c.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		if l.allowRespBody {
			c.Writer = &responseWriter{
				ResponseWriter: c.Writer,
				al:             &al,
			}
		}
		start := time.Now()
		defer func() {
			al.Duration = time.Since(start).String()
			l.logFn(c, al)
		}()

		c.Next()
	}
}

type AccessLog struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	ReqBody  string `json:"req_body"`
	Status   int    `json:"status"`
	RespBody string `json:"resp_body"`
	Duration string `json:"duration"`
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(status int) {
	w.al.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) WriteString(data string) (int, error) {
	w.al.RespBody = data
	return w.ResponseWriter.WriteString(data)
}
