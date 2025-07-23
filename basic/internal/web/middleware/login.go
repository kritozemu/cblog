package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (b *LoginMiddlewareBuilder) IgnorePath(path string) *LoginMiddlewareBuilder {
	b.paths = append(b.paths, path)
	return b
}

func (b *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 将传入的地址忽略掉
		for _, path := range b.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//sess.Options(sessions.Options{
		//	MaxAge: 60,
		//})
		now := time.Now()
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			sess.Set(updateTimeKey, now)
			sess.Set("userId", id)
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
