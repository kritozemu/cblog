package domain

import "time"

// User 个人账户信息
type User struct {
	Id       int64
	Email    string
	Nickname string
	Password string
	Birthday time.Time
	AboutMe  string
	Phone    string

	Ctime time.Time
}
