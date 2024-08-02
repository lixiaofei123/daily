package services

import (
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/repositories"
)

type CommentService interface {
	Create(pid, uid uint, comment *models.Comment) (*models.Comment, error)

	GetByPostId(cuid uint, role string, pid uint, count uint) ([]*models.CommentDetail, error)

	GetById(cid uint) (*models.CommentDetail, error)

	Delete(uid uint, role string, cid uint) error

	SetApproved(uid uint, role string, cid uint) error
}

func NewCommentService(commentRepository repositories.CommentRepository,
	postService PostService,
	userService UserService) CommentService {
	return &commentService{
		commentRepository: commentRepository,
		postService:       postService,
		userService:       userService,
	}
}

type commentService struct {
	commentRepository repositories.CommentRepository
	postService       PostService
	userService       UserService
}

func (cs *commentService) GetByPostId(cuid uint, role string, pid uint, count uint) ([]*models.CommentDetail, error) {

	var comments []*models.Comment
	var err error
	if role == "admin" {
		// 管理员可以查看到所有的评论(包括未审核)
		comments, err = cs.commentRepository.GetByPostId(pid, count)
	} else {
		var post *models.PostDetail
		post, err = cs.postService.Get(cuid, pid)
		if err != nil {
			return nil, err
		}

		if post.Post.UserID == cuid {
			// 如果当前用同时也是这条动态的作者的时候，也可以看到所有的评论(包括未审核)
			comments, err = cs.commentRepository.GetByPostId(pid, count)
		} else {
			// 其余人只能看到已经审核通过的言论
			comments, err = cs.commentRepository.GetByPostIdAndIsApproved(pid, true, count)
		}
	}

	if err != nil {
		return nil, err
	}

	var commentDetails []*models.CommentDetail
	for _, comment := range comments {

		commentDetail := &models.CommentDetail{
			Comment: comment,
		}

		userDetail, _ := cs.userService.GetById(comment.UserID)
		commentDetail.UserDetail = userDetail

		if comment.ReplyTo != nil && comment.ReplyTo.Uid != 0 {
			replyUserDetail, _ := cs.userService.GetById(comment.ReplyTo.Uid)
			commentDetail.ReplyTo = replyUserDetail
		}

		commentDetails = append(commentDetails, commentDetail)
	}

	return commentDetails, nil

}

func (cs *commentService) GetById(cid uint) (*models.CommentDetail, error) {
	comment, err := cs.commentRepository.Get(cid)
	if err != nil {
		return nil, err
	}

	commentDetail := &models.CommentDetail{
		Comment: comment,
	}

	userDetail, _ := cs.userService.GetById(comment.UserID)
	commentDetail.UserDetail = userDetail

	if comment.ReplyTo != nil && comment.ReplyTo.Uid != 0 {
		replyUserDetail, _ := cs.userService.GetById(comment.ReplyTo.Uid)
		commentDetail.ReplyTo = replyUserDetail
	}

	return commentDetail, nil

}

func (cs *commentService) Create(pid, uid uint, comment *models.Comment) (*models.Comment, error) {

	if len(comment.Content) == 0 {
		return nil, apperr.ErrCommentIsEmpty
	}

	comment.ID = 0
	comment.UserID = uid
	comment.PostId = pid

	err := cs.commentRepository.Create(comment)
	if err != nil {
		return nil, err
	}

	return comment, nil

}

func (cs *commentService) Delete(uid uint, role string, cid uint) error {

	comment, err := cs.commentRepository.Get(cid)
	if err != nil {
		return err
	}

	// 管理员可以直接删除
	if role == "admin" {
		return cs.commentRepository.DeleteById(cid)
	}

	// 自己可以删除自己的
	if comment.UserID == uid {
		return cs.commentRepository.DeleteByIdAndUserId(cid, uid)
	}

	// 博主也可以删除
	post, err := cs.postService.Get(uid, comment.PostId)
	if err != nil {
		return err
	}

	if post.Post.UserID == uid {
		return cs.commentRepository.DeleteById(cid)
	}

	return apperr.ErrRoleIsNotEnough

}

func (cs *commentService) SetApproved(uid uint, role string, cid uint) error {

	comment, err := cs.commentRepository.Get(cid)
	if err != nil {
		return err
	}

	// 管理员可以更新
	if role == "admin" {
		comment.IsApproved = true
		return cs.commentRepository.Update(comment)
	}

	// 自己也可以
	post, err := cs.postService.Get(uid, comment.PostId)
	if err != nil {
		return err
	}

	if post.Post.UserID == uid {
		comment.IsApproved = true
		return cs.commentRepository.Update(comment)
	}

	return apperr.ErrRoleIsNotEnough

}
