package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type InitController struct {
	userSrv services.UserService
}

func NewInitController(userSrv services.UserService) *InitController {
	return &InitController{
		userSrv: userSrv,
	}
}

func (au *InitController) GetCheck(ctx echo.Context) mvc.Result {

	userCount, err := au.userSrv.CountUser()
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	result := map[string]interface{}{}
	result["userCount"] = userCount
	return mvc.Ok(result)
}

func (au *InitController) PutRoot(ctx echo.Context, user models.User) mvc.Result {

	userCount, err := au.userSrv.CountUser()
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	if userCount == 0 {
		newuser, err := au.userSrv.AddNewUser(user.Email, user.Password)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		} else {
			return mvc.Ok(newuser)
		}
	} else {
		return mvc.Error(http.StatusBadRequest, apperr.ErrAdminUserIsAlreayAdd)
	}

}
