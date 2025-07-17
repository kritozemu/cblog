package web

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/errs"
	"compus_blog/basic/internal/pkg/ginx"
	"compus_blog/basic/internal/service"
	ijwt "compus_blog/basic/internal/web/jwt"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	svc            service.UserService
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService, hdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		svc:            svc,
		Handler:        hdl,
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapBodyV1(h.SignUp))
	ug.POST("login", ginx.WrapBodyV1(h.LoginJWT))
	// POST /users/edit
	ug.POST("/edit", ginx.WrapBodyAndToken(h.Edit))
	ug.POST("/logout", h.LogoutJwt)
	// GET /users/profile
	ug.GET("/profile", ginx.WrapToken(h.Profile))
	ug.GET("/refresh_token", h.RefreshToken)
}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "非法邮箱格式",
		}, nil
	}
	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入的密码不相等",
		}, nil
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含字母、数字、特殊字符",
		}, nil
	}
	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch {
	case err == nil:
		return ginx.Result{
			Msg: "OK",
		}, nil
	case errors.Is(err, service.ErrDuplicateEmail):
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "邮箱冲突",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}

}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十分钟
			MaxAge: 30,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginJwtReq) (ginx.Result, error) {
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			return ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Msg: "OK",
		}, nil
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或者密码错误",
		}, nil
	default:
		return ginx.Result{Msg: "系统错误"}, err
	}
}

func (h *UserHandler) LogoutJwt(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "退出登录成功"})
}

func (h *UserHandler) Edit(ctx *gin.Context, req UserEditReq,
	uc ijwt.UserClaims) (ginx.Result, error) {
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "生日格式不对",
		}, err
	}
	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
		Id:       uc.Uid,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *UserHandler) Profile(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	u, err := h.svc.FindById(ctx, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Data: ProfileUser{
			Nickname: u.Nickname,
			Email:    u.Email,
			AboutMe:  u.AboutMe,
			Birthday: u.Birthday.Format(time.DateOnly),
		},
	}, nil
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	// 约定，前端在 Authorization 里面带上这个 refresh_token
	tokenStr := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// token 无效或者 redis 有问题
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}
