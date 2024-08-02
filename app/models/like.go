package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Like struct {
	gorm.Model

	UserID uint   `gorm:"column:userId;not null" json:"userId"`
	PostId uint   `gorm:"column:postId;not null;index:index_postid" json:"postId"`
	IP     string `gorm:"column:ip;not null" json:"ip"`
}

func (l *Like) MarshalJSON() ([]byte, error) {
	l.IP = ""
	type Alias Like
	return json.Marshal((*Alias)(l))
}

type LikeDetail struct {
	PostDetail *PostDetail `json:"post"`
	UserDetail *UserDetail `json:"user"`
	Like       *Like       `json:"like"`
}

type LikeDetails struct {
	PostId           uint          `json:"postId"`
	LikeDetails      []*LikeDetail `json:"likeDetails"`
	VisitorLikeCount uint          `json:"visitorLikeCount"`
	IsLiked          bool          `json:"isLiked"`
}
