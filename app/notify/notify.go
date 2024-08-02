package notify

import "github.com/lixiaofei123/daily/app/models"

type NotifyType string

const (
	PostNotifyType    NotifyType = "post"
	CommentNotifyType NotifyType = "comment"
	LikeNotifyType    NotifyType = "like"
)

type EventType string

const (
	CreateEventType EventType = "create"
	UpdateEventType EventType = "update"
	DeleteEventType EventType = "delete"
)

type Notify interface {
	CreatePost(post *models.PostDetail) error

	UpdatePost(post *models.PostDetail) error

	DeletePost(post *models.PostDetail) error

	CreateComment(comment *models.CommentDetail) error

	UpdateComment(comment *models.CommentDetail) error

	DeleteComment(comment *models.CommentDetail) error

	CreateLike(like *models.LikeDetail) error

	UpdateLike(like *models.LikeDetail) error

	DeleteLike(like *models.LikeDetail) error
}
