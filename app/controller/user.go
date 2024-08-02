package controller

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/models"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
)

type UserController struct {
	userSrv services.UserService
}

func NewUserController(userSrv services.UserService) *UserController {
	return &UserController{
		userSrv: userSrv,
	}
}

func (u *UserController) Get(ctx echo.Context) mvc.Result {

	var user *models.UserDetail
	var err error
	id, ok := ctx.Get("id").(uint)
	if !ok {
		user, err = u.userSrv.GetAdmin()
	} else {
		user, err = u.userSrv.GetById(id)
	}

	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(user)
}

func (u *UserController) Post(ctx echo.Context, userDetail models.UserDetail) mvc.Result {

	id := ctx.Get("id").(uint)
	user, err := u.userSrv.UpdateUser(id, &userDetail)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	} else {
		return mvc.Ok(user)
	}
}

func (u *UserController) PostCheck(ctx echo.Context) mvc.Result {
	id := ctx.Get("id").(uint)

	user, err := u.userSrv.GetById(id)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(user)
}

func (u *UserController) GetBy(ctx echo.Context, email string) mvc.Result {

	user, err := u.userSrv.GetByEmail(email)
	if err != nil {
		return mvc.Error(http.StatusInternalServerError, err)
	}

	return mvc.Ok(user)
}
