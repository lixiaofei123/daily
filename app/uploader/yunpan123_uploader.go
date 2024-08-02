package uploader

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"time"

	"math/rand"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/configs"
	"github.com/studio-b12/gowebdav"
)

func init() {
	RegsiterDriver("123YunPan", &YunPan123Uploader{}, &YunPan123UploaderConfig{})
}

type YunPan123Uploader struct {
	client   *gowebdav.Client
	dlZone   string
	signKey  string
	uid      string
	signTime int
}

type YunPan123UploaderConfig struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	DLZone   string `json:"dlzone"`
	Uid      string `json:"uid"`
	SignKey  string `json:"signKey"`
	SignTime string `json:"signTime"`
}

func (y *YunPan123Uploader) initConfig(config interface{}) error {

	uploaderConfig := config.(*YunPan123UploaderConfig)

	client := gowebdav.NewClient(uploaderConfig.Address, uploaderConfig.User, uploaderConfig.Password)
	err := client.Connect()
	if err != nil {
		log.Panicln(err)
	}

	y.client = client
	y.dlZone = uploaderConfig.DLZone
	y.uid = uploaderConfig.Uid
	y.signKey = uploaderConfig.SignKey

	signTimeStr := uploaderConfig.SignTime
	signTime, _ := strconv.Atoi(signTimeStr)
	if signTime <= 0 {
		signTime = 60
	}

	y.signTime = signTime

	return nil

}

func (y *YunPan123Uploader) InitUploader(c *echo.Group) error {
	c.PUT("/yunpan123/*", func(ctx echo.Context) error {
		savepath := ctx.Request().URL.Path[len("/uploader/yunpan123/"):]
		yunpan123_path := path.Join(y.dlZone, savepath)

		body := ctx.Request().Body
		defer body.Close()

		err := y.client.WriteStream(yunpan123_path, body, 0644)
		if err != nil {
			return err
		}

		ctx.JSON(200, nil)

		return nil

	}, checkAuthMiddle("/uploader/yunpan123/"))

	return nil
}

func (y *YunPan123Uploader) GetUploadUrl(path string) (string, error) {

	secret := configs.GlobalConfig.Auth.Secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("/uploader/yunpan123/%s?Authorization=%s", path, authorization), nil
}

func signURL(originURL, privateKey string, uid string, validDuration time.Duration) (newURL string, err error) {

	var (
		ts     = time.Now().Add(validDuration).Unix() // 有效时间戳
		rInt   = rand.Int()                           // 随机正整数
		objURL *url.URL
	)
	objURL, err = url.Parse(originURL)
	if err != nil {
		return "", err
	}
	authKey := fmt.Sprintf("%d-%d-%s-%x", ts, rInt, uid, md5.Sum([]byte(fmt.Sprintf("%s-%d-%d-%s-%s",
		objURL.Path, ts, rInt, uid, privateKey))))
	v := objURL.Query()
	v.Add("auth_key", authKey)
	objURL.RawQuery = v.Encode()
	return objURL.String(), nil
}

func (y *YunPan123Uploader) GetDownloadUrl(path string) (string, error) {
	if y.signKey == "" {
		return fmt.Sprintf("https://vip.123pan.cn/%s/%s/%s", y.uid, y.dlZone, path), nil
	} else {
		originURL := fmt.Sprintf("https://vip.123pan.cn/%s/%s/%s", y.uid, y.dlZone, path)
		newurl, _ := signURL(originURL, y.signKey, y.uid, time.Second*time.Duration(y.signTime))
		return newurl, nil
	}
}

func (y *YunPan123Uploader) Put(path string, data []byte) error {

	err := y.client.Write(path, data, 0644)
	return err
}
