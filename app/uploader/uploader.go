package uploader

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/configs"
)

type Uploader interface {
	initConfig(config interface{}) error

	InitUploader(c *echo.Group) error

	GetUploadUrl(path string) (string, error)

	GetDownloadUrl(path string) (string, error)

	Put(path string, data []byte) error
}

var uploaders map[string]Uploader = map[string]Uploader{}

type UploaderConfig interface {
}

var uploaderConfigs map[string]UploaderConfig = map[string]UploaderConfig{}

func RegsiterDriver(name string, uploader Uploader, uploaderConfig UploaderConfig) {
	uploaders[name] = uploader
	uploaderConfigs[name] = uploaderConfig

}
func NewUploader(name string, config map[string]string) Uploader {

	data, err := json.Marshal(config)
	if err != nil {
		log.Panic(err)
	}

	uploaderConfig, ok := uploaderConfigs[name]
	if !ok {
		log.Panic(err)
	}

	err = json.Unmarshal(data, uploaderConfig)

	uploader, ok := uploaders[name]
	if !ok {
		log.Panic(err)
	}

	err = uploader.initConfig(uploaderConfig)
	if err != nil {
		log.Panic(err)
	}

	return uploader
}

func checkAuthMiddle(prefix string) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authorization := ctx.QueryParam("Authorization")
			token, err := jwt.ParseWithClaims(authorization, &LocalFileSignJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(configs.GlobalConfig.Auth.Secret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*LocalFileSignJWTClaims); ok {
					path := ctx.Request().URL.Path[len(prefix):]
					if claims.Path == path {
						return next(ctx)
					}
				}
			}

			ctx.JSON(http.StatusUnauthorized, mvc.DataResponse{
				Code: http.StatusUnauthorized,
				Data: apperr.ErrSignIsError,
			})
			return nil

		}
	}

}
