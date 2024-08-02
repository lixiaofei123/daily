package models

import (
	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/utils"
	"gorm.io/gorm"
)

type MediaType string

const (
	FileMediaType    MediaType = "file"
	VideoMediaType   MediaType = "video"
	MusicMediaType   MediaType = "music"
	ImageMediaType   MediaType = "image"
	UnknownMediaType MediaType = "unknown"
)

type Media struct {
	Type      MediaType `json:"type"`
	Url       string    `json:"url"`
	Thumbnail string    `json:"thumbnail"`
	Title     *string   `json:"title"`
	Artist    *string   `json:"artist"`
	Album     *string   `json:"ablum"`
	Duration  int64     `json:"duration"`
}

type PostType string

const (
	TextPostType     PostType = "text"
	ImagePostType    PostType = "image"
	FilePostType     PostType = "file"
	UrlPostType      PostType = "url"
	MusicPostType    PostType = "music"
	VideoPostType    PostType = "video"
	ExternalPostType PostType = "external"
	CardPostType     PostType = "card"
)

type Url struct {
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Url      string `json:"url"`
}

type ExternalMedia struct {
	Type   MediaType         `json:"type"`
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

type Card struct {
	Name  string      `json:"name"`
	Model interface{} `json:"model"`
}

type Visibility string

const (
	PublicVisibility  Visibility = "public"
	PrivateVisibility Visibility = "private"
)

type Post struct {
	gorm.Model

	UserID        uint           `gorm:"column:userId;index" json:"userId"`
	PostType      PostType       `gorm:"column:postType;not null" json:"postType"`
	Content       string         `gorm:"column:text;" json:"content"`
	Medias        []Media        `gorm:"column:medias;serializer:json" json:"medias"`
	Url           *Url           `gorm:"column:url;serializer:json" json:"url"`
	ExternalMedia *ExternalMedia `gorm:"column:externalMedia;serializer:json" json:"externalMedia"`
	Card          *Card          `gorm:"column:card;serializer:json" json:"card"`
	IsApproved    bool           `gorm:"column:isApproved;default:false" json:"isApproved"`
	Address       string         `gorm:"column:address;" json:"address"`
	Visibility    Visibility     `gorm:"column:visibility;" json:"visibility"`
	Priority      uint           `gorm:"column:priority;default:0" json:"priority"`
	IsUpdated     bool           `gorm:"column:isUpdated;default:false" json:"isUpdated"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {

	if p.Url != nil && !utils.IsValidHttpURL(p.Url.Url) {
		return apperror.ErrInvalidURL
	}
	if p.ExternalMedia != nil {
		err := FastCheckExternalMedia(p.ExternalMedia)
		if err != nil {
			return err
		}
	}
	return
}

type PostDetail struct {
	UserDetail *UserDetail `json:"user"`
	Post       *Post       `json:"post"`
}
