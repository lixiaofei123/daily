package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
	"github.com/lixiaofei123/daily/app/utils"
)

var secret string
var root string

func init() {
	secret = getEnv("SECRET", "")
	if secret == "" {
		log.Panic("必须要指定环境变量AUTH_SECRET")
	}

	root = getEnv("ROOT", "")
	if root == "" {
		log.Panic("必须要指定环境变量ROOT")
	}
}

type LocalFileSignJWTClaims struct {
	Path string `json:"path"`
	jwt.StandardClaims
}

func checkAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {

		if ctx.Request().Method == "OPTIONS" {
			return next(ctx)
		}

		authorization := ctx.QueryParam("Authorization")
		token, err := jwt.ParseWithClaims(authorization, &LocalFileSignJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*LocalFileSignJWTClaims); ok {
				path := ctx.Request().URL.Path[1:]
				if claims.Path == path {
					return next(ctx)
				}
			}
		}

		ctx.JSON(http.StatusUnauthorized, nil)
		return nil

	}
}

func main() {

	e := echo.New()
	e.Use(echo_middleware.Recover())
	e.Use(echo_middleware.CORSWithConfig(echo_middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowCredentials: true,
	}))

	c := e.Group("")

	c.PUT("/*", func(ctx echo.Context) error {
		savepath := ctx.Request().URL.Path[1:]
		absolutePath, err := utils.SafeJoin(root, savepath)
		if err != nil {
			ctx.String(http.StatusForbidden, err.Error())
			return nil
		}

		dir := path.Dir(absolutePath)
		os.MkdirAll(dir, 0755)

		file, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			return err
		}

		body := ctx.Request().Body
		defer body.Close()

		_, err = io.Copy(file, body)
		if err != nil {
			return err
		}

		ctx.JSON(200, nil)

		return nil

	}, checkAuth)

	c.GET("/*", func(ctx echo.Context) error {

		downloadpath := ctx.Request().URL.Path[1:]

		absolutePath, err := utils.SafeJoin(root, downloadpath)
		if err != nil {
			ctx.String(http.StatusForbidden, err.Error())
			return nil
		}

		_, err = os.Stat(absolutePath)
		if os.IsNotExist(err) {
			ctx.JSON(404, nil)
		}

		ctx.File(absolutePath)
		return nil
	}, checkAuth)

	e.Logger.Fatal(e.Start(":8082"))

}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
