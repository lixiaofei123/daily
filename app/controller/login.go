package controller

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	echo "github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/models"
	mvc "github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/services"
	"github.com/lixiaofei123/daily/configs"
)

type LoginController struct {
	userSrv services.UserService
}

func NewLoginController(userSrv services.UserService) *LoginController {
	return &LoginController{
		userSrv: userSrv,
	}
}

func (u *LoginController) PostLogin(ctx echo.Context, user models.User) mvc.Result {

	loginuser, err := u.userSrv.Login(user.Email, user.Password)
	if err != nil {
		return mvc.Error(http.StatusUnauthorized, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserJWTClaims{
		Role:     string(loginuser.User.Role),
		Email:    loginuser.User.Email,
		Username: loginuser.User.Username,
		Enable:   loginuser.User.Enable,
		UserId:   loginuser.User.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
			Issuer:    user.Username,
		},
	})

	authorization, _ := token.SignedString([]byte(configs.GlobalConfig.Auth.Secret))

	return mvc.Ok(authorization)
}

func (u *LoginController) PostBotLogin(ctx echo.Context, user models.User) mvc.Result {

	loginuser, err := u.userSrv.Login(user.Email, user.Password)
	if err != nil {
		return mvc.Error(http.StatusUnauthorized, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserJWTClaims{
		Role:     string(loginuser.User.Role),
		Email:    loginuser.User.Email,
		Username: loginuser.User.Username,
		Enable:   loginuser.User.Enable,
		UserId:   loginuser.User.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(3650 * 24 * time.Hour).Unix(),
			Issuer:    user.Username,
		},
	})

	authorization, _ := token.SignedString([]byte(configs.GlobalConfig.Auth.Secret))

	return mvc.Ok(authorization)
}
