package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type AdminUserController struct {
	userSrv services.UserService
}

func NewAdminUserController(userSrv services.UserService) *AdminUserController {
	return &AdminUserController{
		userSrv: userSrv,
	}
}

func (au *AdminUserController) Put(ctx echo.Context, user models.User) mvc.Result {

	newuser, err := au.userSrv.AddNewUser(user.Email, user.Password)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(newuser)
}

func (au *AdminUserController) PostEnableBy(ctx echo.Context, uid int) mvc.Result {

	bodyMap, err := mvc.GetMapValueFromBody(ctx)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	enable := GetBoolValueFromMap(bodyMap, "enable")
	if enable != nil {
		user, err := au.userSrv.UpdateUserEnable(uint(uid), *enable)
		if err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		} else {
			return mvc.Ok(user)
		}
	}

	return mvc.Error(http.StatusBadRequest, apperr.ErrInvalidParam)
}

func (au *AdminUserController) PostPasswordBy(ctx echo.Context, uid int) mvc.Result {
	bodyMap, err := mvc.GetMapValueFromBody(ctx)
	if err != nil {
		return mvc.Error(http.StatusBadRequest, err)
	}

	password := GetStringValueFromMap(bodyMap, "password")
	if password != nil {
		if err = au.userSrv.ResetUserPassword(uint(uid), *password); err != nil {
			return mvc.Error(http.StatusInternalServerError, err)
		} else {
			return mvc.Ok("")
		}
	}
	return mvc.Error(http.StatusBadRequest, apperr.ErrInvalidParam)

}

func (au *AdminUserController) GetAll(ctx echo.Context) mvc.Result {
	users, err := au.userSrv.ListUsers()
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}
	return mvc.Ok(users)
}
