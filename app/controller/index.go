package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
	"github.com/lixiaofei123/daily/configs"
)

type IndexController struct {
	postSrv services.PostService
	userSrv services.UserService
}

func NewIndexController(postSrv services.PostService, userSrv services.UserService) *IndexController {
	return &IndexController{
		postSrv: postSrv,
		userSrv: userSrv,
	}
}

func (i *IndexController) Get(ctx echo.Context) mvc.Result {

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	userDetail, err := i.userSrv.GetUserOrAdmin(id)
	if err != nil {
		userDetail = &models.UserDetail{
			User:    &models.User{},
			Profile: &models.UserProfile{},
		}
	}

	posts, err := i.postSrv.List(id, 1, 20)
	if err != nil {
		posts = &models.PageResult{
			Page:     0,
			PageSize: 0,
			Data:     []*models.PostDetail{},
		}
	}

	data := map[string]interface{}{}
	data["user"] = userDetail
	data["posts"] = posts
	data["cuid"] = id
	data["isDetail"] = false
	data["isMobile"] = IsMobileUserAgent(ctx)

	return mvc.OkHtml("index.html", data)
}

func (i *IndexController) GetPageBy(ctx echo.Context, page int) mvc.Result {

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	posts, err := i.postSrv.List(id, uint(page), 20)
	if err != nil {
		return mvc.ErrorHtml(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{}
	data["posts"] = posts
	data["cuid"] = id
	data["isMobile"] = IsMobileUserAgent(ctx)

	return mvc.OkHtml("page.html", data)
}

func (i *IndexController) GetUserBy(ctx echo.Context, userid int) mvc.Result {

	site := configs.GlobalConfig.Site

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	userDetail, err := i.userSrv.GetById(uint(userid))
	if err != nil {
		return mvc.Error(http.StatusNotFound, apperr.ErrUserNotFound)
	}

	if userDetail.User.Role == models.VisitorRole {
		return mvc.Error(http.StatusNotFound, apperr.ErrUserNotFound)
	}

	posts, err := i.postSrv.ListByUid(id, uint(userid), 1, 20)
	if err != nil {
		posts = &models.PageResult{
			Page:     0,
			PageSize: 0,
			Data:     []*models.PostDetail{},
		}
	}

	data := map[string]interface{}{}
	data["user"] = userDetail
	data["posts"] = posts
	data["site"] = site
	data["cuid"] = id
	data["uid"] = userid
	data["isDetail"] = false
	data["isMobile"] = IsMobileUserAgent(ctx)

	return mvc.OkHtml("index.html", data)
}

func (i *IndexController) GetUserPageBy(ctx echo.Context, userid int, page int) mvc.Result {

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	posts, err := i.postSrv.ListByUid(id, uint(userid), uint(page), 20)
	if err != nil {
		return mvc.ErrorHtml(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{}
	data["posts"] = posts
	data["cuid"] = id
	data["uid"] = userid
	data["isMobile"] = IsMobileUserAgent(ctx)

	return mvc.OkHtml("page.html", data)
}

func (i *IndexController) GetPostBy(ctx echo.Context, pid int) mvc.Result {

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	post, err := i.postSrv.Get(id, uint(pid))
	if err != nil {
		return mvc.ErrorHtml(http.StatusNotFound, apperr.ErrPostNotFound)
	}

	data := map[string]interface{}{}
	data["user"] = post.UserDetail
	data["posts"] = &models.PageResult{
		Page:     1,
		PageSize: 1,
		Data:     []*models.PostDetail{post},
	}
	data["isDetail"] = true
	data["cuid"] = id
	data["isMobile"] = IsMobileUserAgent(ctx)

	return mvc.OkHtml("index.html", data)
}
