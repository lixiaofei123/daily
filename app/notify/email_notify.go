package notify

import (
	"fmt"
	"net/smtp"
	"strconv"

	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
)

type EmailNotify struct {
	auth       smtp.Auth
	sendEmail  string
	recvEmail  string
	smtpServer string
	smtpPort   uint
}

type EmailNotifyConfig struct {
	Email      string `json:"sendEmail"`
	RecvEmail  string `json:"recvEmail"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	SMTPServer string `json:"smtpServer"`
	SMTPPort   string `json:"smtpPort"`
}

func (e *EmailNotify) init(config EmailNotifyConfig) {
	auth := smtp.PlainAuth("", config.Username,
		config.Password, config.SMTPServer)
	e.auth = auth
	e.sendEmail = config.Email
	e.recvEmail = config.RecvEmail
	e.smtpServer = config.SMTPServer
	smtpPort, _ := strconv.Atoi(config.SMTPPort)
	if smtpPort <= 0 {
		smtpPort = 25
	}
	e.smtpPort = uint(smtpPort)
}

func (e *EmailNotify) CreatePost(post *models.PostDetail) error {

	text := fmt.Sprintf(`通知: %s 发布了一篇新的动态，内容是 %s`,
		post.UserDetail.User.Username, post.Post.Content)

	return smtp.SendMail(fmt.Sprintf("%s:%d", e.smtpServer, e.smtpPort),
		e.auth, e.sendEmail, []string{e.recvEmail}, []byte(text))
}

func (e *EmailNotify) UpdatePost(post *models.PostDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) DeletePost(post *models.PostDetail) error {
	text := fmt.Sprintf(`通知: %s 发布了动态被删除，内容是 %s`,
		post.UserDetail.User.Username, post.Post.Content)

	return smtp.SendMail(fmt.Sprintf("%s:%d", e.smtpServer, e.smtpPort),
		e.auth, e.sendEmail, []string{e.recvEmail}, []byte(text))
}

func (e *EmailNotify) CreateComment(comment *models.CommentDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) UpdateComment(comment *models.CommentDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) DeleteComment(comment *models.CommentDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) CreateLike(like *models.LikeDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) UpdateLike(like *models.LikeDetail) error {
	return apperr.ErrNotYetImplementedMethod
}

func (e *EmailNotify) DeleteLike(like *models.LikeDetail) error {
	return apperr.ErrNotYetImplementedMethod
}
