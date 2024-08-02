package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type LBSController struct {
	service services.LBSService
}

func NewLBSController(service services.LBSService) *LBSController {
	return &LBSController{
		service: service,
	}
}

func (lc *LBSController) GetMypos(ctx echo.Context) mvc.Result {

	ip := ctx.RealIP()
	address, err := lc.service.GetAddressByIp(ip)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(address)
}
