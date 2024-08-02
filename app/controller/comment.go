package controller

import (
	"net/http"
	"unicode/utf8"

	echo "github.com/labstack/echo/v4"
	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type CommentController struct {
	commentSrv services.CommentService
	likeSrv    services.LikeService
	userSrv    services.UserService
	postSrv    services.PostService
}

func NewCommentController(commentSrv services.CommentService, likeSrv services.LikeService, userSrv services.UserService, postSrv services.PostService) *CommentController {
	return &CommentController{
		commentSrv: commentSrv,
		userSrv:    userSrv,
		postSrv:    postSrv,
		likeSrv:    likeSrv,
	}
}

func (cc *CommentController) Put(ctx echo.Context, comment models.Comment) mvc.Result {

	realIp := ctx.RealIP()
	comment.IP = realIp

	if comment.Content == "" {
		return mvc.Error(http.StatusBadRequest, apperror.ErrCommentIsEmpty)
	}

	commentLength := utf8.RuneCountInString(comment.Content)

	if commentLength > 100 {
		return mvc.Error(http.StatusBadRequest, apperror.ErrCommentIsTooLong)
	}

	cuid, ok := ctx.Get("id").(uint)

	post, err := cc.postSrv.Get(cuid, comment.PostId)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	if ok {
		comment.IsApproved = true
		comment, err := cc.commentSrv.Create(post.Post.ID, cuid, &comment)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}

		commentDetail, err := cc.commentSrv.GetById(comment.ID)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}

		return mvc.Ok(commentDetail)
	} else {
		// 用户需要传入邮箱
		email := mvc.GetValue(ctx, "email")
		username := mvc.GetValue(ctx, "username")

		if email == "" || username == "" {
			return mvc.Error(http.StatusBadRequest, apperror.ErrRequiredParamMissing)
		}

		user, err := cc.userSrv.CreateOrGetByEmail(email, username)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}

		comment.IsApproved = false
		comment, err := cc.commentSrv.Create(post.Post.ID, user.ID, &comment)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}

		commentDetail, err := cc.commentSrv.GetById(comment.ID)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}

		return mvc.Ok(commentDetail)

	}
}

func (cc *CommentController) DeleteBy(ctx echo.Context, cid int) mvc.Result {
	cuid := ctx.Get("id").(uint)
	role := ctx.Get("role").(string)
	err := cc.commentSrv.Delete(cuid, role, uint(cid))
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}
	return mvc.Ok("")
}

func (cc *CommentController) PostApproveBy(ctx echo.Context, cid int) mvc.Result {
	cuid := ctx.Get("id").(uint)
	role := ctx.Get("role").(string)
	err := cc.commentSrv.SetApproved(cuid, role, uint(cid))
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}
	return mvc.Ok("")
}

func (cc *CommentController) GetPostBy(ctx echo.Context, pid int) mvc.Result {
	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}
	role, ok := ctx.Get("role").(string)
	if !ok {
		role = string(models.VisitorRole)
	}

	loadAll := mvc.GetBooleanValueWithDefault(ctx, "loadAll", false)

	var queryCommentCount uint = 6
	if loadAll {
		queryCommentCount = 10000
	}

	comments, err := cc.commentSrv.GetByPostId(cuid, role, uint(pid), queryCommentCount)

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	likes, err := cc.likeSrv.GetByPostId(uint(pid))
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	realIp := ctx.RealIP()
	like, err := cc.likeSrv.CheckIsLiked(realIp, cuid, uint(pid))
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	likes.IsLiked = like != nil

	resut := map[string]interface{}{}
	resut["comments"] = comments
	resut["likes"] = likes

	return mvc.Ok(resut)
}
