package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/card"
	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
	"github.com/lixiaofei123/daily/app/utils"
)

type PostController struct {
	postSrv   services.PostService
	uploadSrv services.UploadService
	userSrv   services.UserService
}

func NewPostController(postSrv services.PostService, uploadSrv services.UploadService, userSrv services.UserService) *PostController {
	return &PostController{
		postSrv:   postSrv,
		uploadSrv: uploadSrv,
		userSrv:   userSrv,
	}
}

func (p *PostController) Put(ctx echo.Context, post models.Post) mvc.Result {

	cuid := ctx.Get("id").(uint)
	post.UserID = cuid

	post.ID = 0
	err := p.postSrv.Create(&post)

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok("")

}

func (p *PostController) Post(ctx echo.Context, post models.Post) mvc.Result {

	cuid := ctx.Get("id").(uint)
	post.UserID = cuid

	existPost, err := p.postSrv.Get(cuid, post.ID)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	if existPost.UserDetail.ID != cuid {
		return mvc.Error(http.StatusForbidden, apperror.ErrRoleIsNotEnough)
	}

	post.CreatedAt = existPost.Post.CreatedAt
	err = p.postSrv.Create(&post)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok("")

}

type BotPost struct {
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Address string   `json:"address"`
}

// 专门为机器人增加的发布图文消息的接口
func (p *PostController) PutBot(ctx echo.Context, botPost BotPost) mvc.Result {

	cuid := ctx.Get("id").(uint)

	post := &models.Post{
		UserID:  cuid,
		Content: botPost.Content,
		Address: botPost.Address,
	}

	if len(botPost.Images) > 0 {
		medias := []models.Media{}
		for _, image := range botPost.Images {
			data, err := utils.HttpGet(image)
			if err != nil {
				return mvc.Error(http.StatusInternalServerError, err)
			}
			path := utils.GenerateLocalPath(image)
			err = p.uploadSrv.PutData(path, data)
			if err != nil {
				return mvc.Error(http.StatusInternalServerError, err)
			}
			url := fmt.Sprintf("/api/v1/file/%s", path)
			medias = append(medias, models.Media{
				Type:      models.ImageMediaType,
				Thumbnail: url,
				Url:       url,
			})
		}
		post.Medias = medias
	}

	err := p.postSrv.Create(post)

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok("")

}

func (p *PostController) Get(ctx echo.Context) mvc.Result {
	page := mvc.GetIntValueWithDefault(ctx, "page", 0)
	pageSize := mvc.GetIntValueWithDefault(ctx, "pageSize", 20)

	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}

	result, err := p.postSrv.List(cuid, uint(page), uint(pageSize))

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(result)
}

func (p *PostController) GetExternalmediaConfig(ctx echo.Context) mvc.Result {
	supportPlatforms := models.GetSupportPlatforms()
	return mvc.Ok(supportPlatforms)
}

func (p *PostController) GetExternalmediaPreview(ctx echo.Context) mvc.Result {

	config := ctx.QueryParams().Get("config")
	externalMedia := new(models.ExternalMedia)
	err := json.Unmarshal([]byte(config), externalMedia)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	externalPlatform, err := models.GetExternalPlatform(externalMedia)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	html := externalPlatform.Html(IsMobileUserAgent(ctx))
	return mvc.Ok(html)
}

func (p *PostController) GetBy(ctx echo.Context, pid int) mvc.Result {

	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}

	result, err := p.postSrv.Get(cuid, uint(pid))

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(result)
}

func (p *PostController) GetUserBy(ctx echo.Context, uid int) mvc.Result {

	page := mvc.GetIntValueWithDefault(ctx, "page", 0)
	pageSize := mvc.GetIntValueWithDefault(ctx, "pageSize", 20)

	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}

	result, err := p.postSrv.ListByUid(cuid, uint(uid), uint(page), uint(pageSize))

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(result)
}

func (p *PostController) DeleteBy(ctx echo.Context, pid int) mvc.Result {

	cuid := ctx.Get("id").(uint)

	err := p.postSrv.DeleteByIdAndUserId(uint(pid), cuid)

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok("")
}

func (p *PostController) PostTopBy(ctx echo.Context, pid int) mvc.Result {

	cuid := ctx.Get("id").(uint)

	role, ok := ctx.Get("role").(string)
	if !ok {
		role = "user"
	}

	priority := 10
	if role == "admin" {
		priority = 20
	}

	result, err := p.postSrv.UpdatePriority(uint(pid), cuid, uint(priority))

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(result)
}

func (p *PostController) DeleteTopBy(ctx echo.Context, pid int) mvc.Result {

	cuid := ctx.Get("id").(uint)

	result, err := p.postSrv.UpdatePriority(uint(pid), cuid, 0)

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(result)
}

func (p *PostController) GetCardConfig(ctx echo.Context) mvc.Result {
	cardModels := card.CardModels
	return mvc.Ok(cardModels)
}

func (p *PostController) GetCardPreview(ctx echo.Context) mvc.Result {
	modelstr := ctx.QueryParams().Get("model")
	name := ctx.QueryParams().Get("name")
	content := ctx.QueryParams().Get("content")

	modeldata := new(map[string]interface{})
	err := json.Unmarshal([]byte(modelstr), modeldata)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	id, ok := ctx.Get("id").(uint)
	if !ok {
		id = 0
	}

	userDetail, _ := p.userSrv.GetById(uint(id))

	postDetail := &models.PostDetail{
		UserDetail: userDetail,
		Post: &models.Post{
			Content: content,
			Address: "模拟地名",
		},
	}
	postDetail.Post.ID = 9999
	postDetail.Post.CreatedAt = time.Now()
	html, err := card.Html(name, modeldata, postDetail)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	return mvc.Ok(html)
}
