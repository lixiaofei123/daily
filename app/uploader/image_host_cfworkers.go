package uploader

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	apperror "github.com/lixiaofei123/daily/app/errors"
)

func init() {
	RegsiterDriver("ImageHostCFWorkers", &ImageHostingCFWorkersUploader{}, &ImageHostingCFWorkersUploaderConfig{})
}

type ImageHostingCFWorkersUploader struct {
	url      string
	signkey  string
	r2domain string
}

type ImageHostingCFWorkersUploaderConfig struct {
	Url      string `json:"url"`
	SignKey  string `json:"signkey"`
	R2Domain string `json:"r2domain"`
}

func (u *ImageHostingCFWorkersUploader) initConfig(config interface{}) error {
	uploaderConfig := config.(*ImageHostingCFWorkersUploaderConfig)

	u.signkey = uploaderConfig.SignKey
	u.url = uploaderConfig.Url
	u.r2domain = uploaderConfig.R2Domain

	return nil

}

func (u *ImageHostingCFWorkersUploader) InitUploader(c *echo.Group) error {
	return nil
}

func (u *ImageHostingCFWorkersUploader) GetUploadUrl(path string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})
	authorization, _ := token.SignedString([]byte(u.signkey))
	return fmt.Sprintf("%s/file/%s?Authentication=Bearer %s", u.url, path, authorization), nil
}

func (u *ImageHostingCFWorkersUploader) GetDownloadUrl(path string) (string, error) {

	if u.r2domain != "" {
		return fmt.Sprintf("%s/%s", u.r2domain, path), nil
	}

	return fmt.Sprintf("%s/file/%s", u.url, path), nil
}

func (u *ImageHostingCFWorkersUploader) Put(path string, data []byte) error {

	uploadUrl, err := u.GetUploadUrl(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", uploadUrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		return nil
	}

	return apperror.ErrUploadError
}
