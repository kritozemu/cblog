package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱或手机号冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindById(ctx context.Context, uid int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
}

type UserDAOStruct struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &UserDAOStruct{db: db}
}

func (dao *UserDAOStruct) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突，邮箱冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAOStruct) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDAOStruct) UpdateById(ctx context.Context, entity User) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).
		Updates(map[string]any{
			"utime":    now,
			"nickname": entity.Nickname,
			"birthday": entity.Birthday,
			"about_me": entity.AboutMe,
		}).Error
	return err
}

func (dao *UserDAOStruct) FindById(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&u).Error
	return u, err
}

func (dao *UserDAOStruct) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 代表这是一个可以为 NULL 的列
	//Email    *string
	Email    sql.NullString `gorm:"unique"`
	Password string

	Nickname string `gorm:"type=varchar(128)"`
	// YYYY-MM-DD
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`

	// 代表这是一个可以为 NULL 的列
	Phone sql.NullString `gorm:"unique"`

	// 时区，UTC 0 的毫秒数
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64

	// json 存储
	//Addr string
}
