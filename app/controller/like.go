package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type LikeController struct {
	likeSrv services.LikeService
}

func NewLikeController(likeSrv services.LikeService) *LikeController {
	return &LikeController{
		likeSrv: likeSrv,
	}
}

func (lc *LikeController) PutPostBy(ctx echo.Context, pid int) mvc.Result {

	cuid, ok := ctx.Get("id").(uint)
	var like *models.Like
	var err error
	realIp := ctx.RealIP()
	if ok {
		like, err = lc.likeSrv.CreateByUid(cuid, uint(pid))
	} else {
		cuid = 0
		like, err = lc.likeSrv.CreateByIp(realIp, uint(pid))
	}

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(like)
}

func (lc *LikeController) DeletePostBy(ctx echo.Context, pid int) mvc.Result {
	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}
	realIp := ctx.RealIP()

	like, err := lc.likeSrv.CheckIsLiked(realIp, cuid, uint(pid))

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	if like != nil {
		err := lc.likeSrv.Delete(like.ID)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		}
	}

	return mvc.Ok("")
}

func (lc *LikeController) GetPostBy(ctx echo.Context, pid int) mvc.Result {
	likes, err := lc.likeSrv.GetByPostId(uint(pid))
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	cuid, ok := ctx.Get("id").(uint)
	if !ok {
		cuid = 0
	}
	realIp := ctx.RealIP()

	like, err := lc.likeSrv.CheckIsLiked(realIp, cuid, uint(pid))
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	likes.IsLiked = like != nil

	return mvc.Ok(likes)
}
