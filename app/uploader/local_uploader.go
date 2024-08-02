package uploader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/utils"
	"github.com/lixiaofei123/daily/configs"
)

func init() {
	RegsiterDriver("Local", &LocalUploader{}, &LocalUploaderConfig{})
}

type LocalUploader struct {
	root       string
	fileserver string
}

type LocalUploaderConfig struct {
	Root       string `json:"root"`
	Fileserver string `json:"fileserver"`
}

type LocalFileSignJWTClaims struct {
	Path string `json:"path"`
	jwt.StandardClaims
}

func (u *LocalUploader) initConfig(config interface{}) error {
	uploaderConfig := config.(*LocalUploaderConfig)
	u.root = uploaderConfig.Root
	u.fileserver = uploaderConfig.Fileserver
	return nil

}

func (u *LocalUploader) InitUploader(c *echo.Group) error {
	c.PUT("/local/*", func(ctx echo.Context) error {
		savepath := ctx.Request().URL.Path[len("/uploader/local/"):]
		absolutePath, err := utils.SafeJoin(u.root, savepath)
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

	}, checkAuthMiddle("/uploader/local/"))

	if u.fileserver == "" {
		c.GET("/local/*", func(ctx echo.Context) error {

			downloadpath := ctx.Request().URL.Path[len("/uploader/local/"):]
			absolutePath, err := utils.SafeJoin(u.root, downloadpath)
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
		}, checkAuthMiddle("/uploader/local/"))
	}

	return nil
}

func (u *LocalUploader) GetUploadUrl(path string) (string, error) {

	secret := configs.GlobalConfig.Auth.Secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("/uploader/local/%s?Authorization=%s", path, authorization), nil
}

func (u *LocalUploader) GetDownloadUrl(path string) (string, error) {

	if u.fileserver != "" {
		return fmt.Sprintf("%s/%s", u.fileserver, path), nil
	}

	secret := configs.GlobalConfig.Auth.Secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("/uploader/local/%s?Authorization=%s", path, authorization), nil
}

func (u *LocalUploader) Put(savepath string, data []byte) error {

	absolutePath := path.Join(u.root, savepath)
	dir := path.Dir(absolutePath)
	os.MkdirAll(dir, 0755)

	file, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}
