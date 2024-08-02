package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/controller"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/configs"
)

func AuthHandler(next echo.HandlerFunc, allowRoles []models.Role, justTest bool) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		authorization := ctx.Request().Header.Get("authorization")
		if authorization == "" {
			cookie, err := ctx.Request().Cookie("Authorization")
			if err == nil {
				authorization = cookie.Value
			}
		}
		token, err := jwt.ParseWithClaims(authorization, &controller.UserJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(configs.GlobalConfig.Auth.Secret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*controller.UserJWTClaims); ok {
				ctx.Set("email", claims.Email)
				ctx.Set("role", claims.Role)
				ctx.Set("id", claims.UserId)

				if justTest {
					return next(ctx)
				}

				userRole := models.Role(claims.Role)
				checkRole := true
				if allowRoles != nil {
					checkRole = false
					for _, role := range allowRoles {
						if role == userRole {
							checkRole = true
							break
						}
					}
				}

				if checkRole {
					return next(ctx)
				} else {
					ctx.JSON(http.StatusUnauthorized, mvc.DataResponse{
						Code: http.StatusUnauthorized,
						Data: apperr.ErrRoleIsNotEnough.Error(),
					})
					return nil
				}

			}
		} else {
			if justTest {
				return next(ctx)
			}
		}

		ctx.JSON(http.StatusUnauthorized, mvc.DataResponse{
			Code: http.StatusUnauthorized,
			Data: apperr.ErrNeedLogin.Error(),
		})
		return nil
	}

}

func UserAuthHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method
		if method == "GET" || method == "HEAD" {
			return AuthHandler(next, nil, true)(c)
		} else {
			return AuthHandler(next, []models.Role{models.AdminRole, models.UserRole}, false)(c)
		}
	}
}

func AdminAuthHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return AuthHandler(next, []models.Role{models.AdminRole}, false)
}

func CommentAuthHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method
		if method == "GET" || method == "PUT" {
			return AuthHandler(next, nil, true)(c)
		} else {
			return AuthHandler(next, []models.Role{models.AdminRole, models.UserRole}, false)(c)
		}
	}
}

func IndexAuthHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return AuthHandler(next, nil, true)(c)
	}
}
