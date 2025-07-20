package web

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type UserEditReq struct {
	// 暂时不支持更改手机号、邮箱、密码
	Nickname string `json:"nickname"`

	// YYYY-MM-DD
	Birthday string `json:"birthday"`
	AboutMe  string `json:"aboutMe"`
}

type LoginJwtReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ProfileUser struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	AboutMe  string `json:"aboutMe"`
	Birthday string `json:"birthday"`
}

type SendLoginSmsReq struct {
	Phone string `json:"phone"`
}

type LoginSMS struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type Result struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data string `json:"data"`
}
