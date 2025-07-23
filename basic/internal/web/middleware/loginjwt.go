package middleware

import (
	ijwt "compus_blog/basic/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type LoginJwtMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJwtMiddlewareBuilder(hdl ijwt.Handler) *LoginJwtMiddlewareBuilder {
	return &LoginJwtMiddlewareBuilder{
		Handler: hdl,
	}
}

func (b *LoginJwtMiddlewareBuilder) IgnorePath(path string) *LoginJwtMiddlewareBuilder {
	b.paths = append(b.paths, path)
	return b
}

func (b *LoginJwtMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range b.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		tokenStr := b.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			// token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			// token是伪造的或者token过期了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = b.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token无效或者redis有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("users", uc)
	}
}
