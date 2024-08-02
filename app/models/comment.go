package models

import (
	"gorm.io/gorm"
)

type ReplyTo struct {
	Uid uint `json:"uid"`
	Cid uint `json:"cid"`
}

type Comment struct {
	gorm.Model

	UserID     uint     `gorm:"column:userId" json:"userId"`
	PostId     uint     `gorm:"column:postId;not null;index" json:"postId"`
	IP         string   `gorm:"column:ip;not null" json:"-"`
	Content    string   `gorm:"column:content;not null" json:"content"`
	ReplyTo    *ReplyTo `gorm:"column:replyTo;serializer:json" json:"replyTo"`
	IsApproved bool     `gorm:"column:isApproved;default:false" json:"isApproved"`
}

type CommentDetail struct {
	PostDetail *PostDetail `json:"post"`
	UserDetail *UserDetail `json:"user"`
	Comment    *Comment    `json:"comment"`
	ReplyTo    *UserDetail `json:"replyto"`
}
