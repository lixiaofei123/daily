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
	RegsiterDriver("Remote", &RemoteUploader{}, &RemoteUploaderConfig{})
}

type RemoteUploader struct {
	secret     string
	url        string
	fileserver string
}

type RemoteUploaderConfig struct {
	Url        string `json:"url"`
	Secret     string `json:"secret"`
	Fileserver string `json:"fileserver"`
}

func (u *RemoteUploader) initConfig(config interface{}) error {
	uploaderConfig := config.(*RemoteUploaderConfig)
	u.secret = uploaderConfig.Secret
	u.url = uploaderConfig.Url
	u.fileserver = uploaderConfig.Fileserver
	return nil

}

func (u *RemoteUploader) InitUploader(c *echo.Group) error {

	return nil
}

func (u *RemoteUploader) GetUploadUrl(path string) (string, error) {

	secret := u.secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("%s/%s?Authorization=%s", u.url, path, authorization), nil
}

func (u *RemoteUploader) GetDownloadUrl(path string) (string, error) {

	if u.fileserver != "" {
		return fmt.Sprintf("%s/%s", u.fileserver, path), nil
	}

	secret := u.secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("%s/%s?Authorization=%s", u.url, path, authorization), nil
}

func (u *RemoteUploader) Put(path string, data []byte) error {

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
