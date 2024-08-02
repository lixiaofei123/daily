package controller

import (
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"
	apperror "github.com/lixiaofei123/daily/app/errors"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type FileController struct {
	uploaderSrv services.UploadService
}

func NewFileController(uploaderSrv services.UploadService) *FileController {
	return &FileController{
		uploaderSrv: uploaderSrv,
	}
}

func (f *FileController) RegisterRouter(routers mvc.ExtraRouter) {
	routers.AddRouter("Post", "/:path", "UploadFile")
	routers.AddRouter("Get", "/:path", "DownloadFile")
}

func (f *FileController) UploadFile(ctx echo.Context) mvc.Result {
	path := ctx.Request().URL.Path[len("/api/v1/file/"):]

	uploadUrl, err := f.uploaderSrv.GetUploadUrl(ctx, path)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, apperror.ErrUploadError)
	}

	return mvc.Ok(uploadUrl)
}

func (f *FileController) DownloadFile(ctx echo.Context) mvc.Result {
	path := ctx.Request().URL.Path[len("/api/v1/file/"):]
	downloadUrl, err := f.uploaderSrv.GetDownloadUrl(ctx, path)
	if err != nil {
		if errors.Is(err, apperror.ErrTooManyRequests) {
			return mvc.Error(http.StatusTooManyRequests, err)
		} else {
			return mvc.Error(http.StatusInternalServerError, apperror.ErrDownloadError)
		}

	}

	ctx.Response().Status = http.StatusFound
	ctx.Response().Header().Set("Location", downloadUrl)
	return mvc.Ok("ok")
}
