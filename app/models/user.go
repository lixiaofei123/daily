package models

import (
	"encoding/json"

	"net/mail"

	apperror "github.com/lixiaofei123/daily/app/errors"
	"gorm.io/gorm"
)

type Role string

const (
	AdminRole   Role = "admin"
	UserRole    Role = "user"
	VisitorRole Role = "visitor"
)

type User struct {
	gorm.Model

	Email    string `gorm:"uniqueIndex;not null;size:30" json:"email"`
	Role     Role   `gorm:"role;not null" json:"role"`
	Username string `gorm:"column:uname;not null;size:20" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Enable   bool   `gorm:"not null;default:true" json:"enable"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {

	if u.Email == "" {
		return apperror.ErrInvalidEmail
	}
	_, err = mail.ParseAddress(u.Email)
	return err
}

func (u *User) MarshalJSON() ([]byte, error) {
	u.Password = ""
	type Alias User
	return json.Marshal((*Alias)(u))
}

type UserDetail struct {
	ID      uint         `json:"id"`
	User    *User        `json:"user"`
	Profile *UserProfile `json:"profile"`
}

type SimpleUser struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Enable   bool   `json:"enable"`
}
