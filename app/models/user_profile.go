package models

import (
	"gorm.io/gorm"
)

type UserProfile struct {
	gorm.Model

	UserID    uint   `gorm:"column:userId;uniqueIndex;not null"  json:"userId"`
	Avatar    string `gorm:"column:avatar"   json:"avatar"`
	Signature string `gorm:"column:signature;size:50"   json:"signature"`
	Cover     string `gorm:"column:cover"   json:"cover"`
}
