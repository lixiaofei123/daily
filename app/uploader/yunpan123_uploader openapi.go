package uploader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/configs"
	"github.com/robfig/cron/v3"
)

func init() {
	RegsiterDriver("123YunPanOpenAPI", &YunPan123OpenAPIUploader{}, &YunPan123OpenAPIUploaderConfig{})
}

type YunPan123OpenAPIUploader struct {
	client   *yunpan123Client
	dlZone   string
	signKey  string
	uid      string
	signTime int
}

type YunPan123OpenAPIUploaderConfig struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	DlZone       string `json:"dlzone"`
	Uid          string `json:"uid"`
	SignKey      string `json:"signKey"`
	SignTime     string `json:"signTime"`
}

func (y *YunPan123OpenAPIUploader) initConfig(config interface{}) error {

	uploaderConfig := config.(*YunPan123OpenAPIUploaderConfig)

	client := newYunpan123Client(uploaderConfig.ClientID, uploaderConfig.ClientSecret)
	err := client.requestAccessToken()
	if err != nil {
		return err
	}

	_, err = client.FindFileByName(0, uploaderConfig.DlZone, 10)
	if err != nil {
		return err
	}

	y.client = client
	y.dlZone = uploaderConfig.DlZone
	y.uid = uploaderConfig.Uid
	y.signKey = uploaderConfig.SignKey

	signTimeStr := uploaderConfig.SignTime
	signTime, _ := strconv.Atoi(signTimeStr)
	if signTime <= 0 {
		signTime = 60
	}

	y.signTime = signTime

	err = y.createTodayDirs()
	if err != nil {
		return err
	}

	cron.New().AddFunc("0 0 * * *", func() {
		y.createTodayDirs()
	})

	return nil
}

// 创建今日的目录，触发缓存
func (y *YunPan123OpenAPIUploader) createTodayDirs() error {
	todaypath := fmt.Sprintf("%s/storage/%s", y.dlZone, time.Now().Format("2006/01/02"))
	_, err := y.client.createDirs(todaypath)
	if err != nil {
		return err
	}
	return nil
}

func (y *YunPan123OpenAPIUploader) InitUploader(c *echo.Group) error {
	c.PUT("/yunpan123_openai/*", func(ctx echo.Context) error {
		savepath := ctx.Request().URL.Path[len("/uploader/yunpan123_openai/"):]
		yunpan123_path := path.Join(y.dlZone, savepath)

		dir := path.Dir(yunpan123_path)
		filename := path.Base(yunpan123_path)

		fileid, err := y.client.createDirs(dir)
		if err != nil {
			return err
		}

		md5 := ctx.QueryParam("md5")
		sizestr := ctx.QueryParam("size")
		size, err := strconv.Atoi(sizestr)
		if err != nil {
			return err
		}

		createResult, err := y.client.createFile(fileid, filename, md5, size, 10)
		if err != nil {
			return err
		}

		data, _ := json.Marshal(createResult)

		ctx.Response().Status = 200
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().Write(data)

		return nil

	}, checkAuthMiddle("/uploader/yunpan123_openai/"))

	c.GET("/yunpan123_openai/*", func(ctx echo.Context) error {

		preuploadID := ctx.QueryParam("preuploadID")
		sliceNoStr := ctx.QueryParam("sliceNo")
		sliceNo, err := strconv.Atoi(sliceNoStr)
		if err != nil {
			return err
		}

		getPreUploadResult, err := y.client.getPreUploadUrl(preuploadID, sliceNo, 10)
		if err != nil {
			return err
		}

		data, _ := json.Marshal(getPreUploadResult)

		ctx.Response().Status = 200
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().Write(data)

		return nil

	}, checkAuthMiddle("/uploader/yunpan123_openai/"))

	c.POST("/yunpan123_openai/*", func(ctx echo.Context) error {

		body := ctx.Request().Body
		defer body.Close()
		bodydata, err := io.ReadAll(body)
		if err != nil {
			return err
		}

		md5s := []string{}
		err = json.Unmarshal(bodydata, &md5s)
		if err != nil {
			return err
		}

		preuploadID := ctx.QueryParam("preuploadID")

		parts, err := y.client.listUploadParts(preuploadID, 10)
		if err != nil {
			return err
		}

		if len(parts.Parts) > 0 {
			if len(parts.Parts) == len(md5s) {
				for i := 0; i < len(parts.Parts); i++ {
					if parts.Parts[i].Etag != md5s[i] {
						return apperror.ErrUnknowPan123Error
					}
				}
			} else {
				return apperror.ErrUnknowPan123Error
			}
		}

		// 数据校验通过，通知完成
		uploadResult, err := y.client.uploadComplete(preuploadID, 10)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		if uploadResult.Async {
			for {
				resp, err := y.client.uploadAsyncResult(preuploadID, 10)
				if err != nil {
					log.Println(err.Error())
					return err
				}
				if resp.Completed {
					ctx.Response().Write([]byte("ok"))
					break
				} else {
					time.Sleep(time.Second)
				}
			}

		} else {
			ctx.Response().Write([]byte("ok"))
		}

		return nil

	}, checkAuthMiddle("/uploader/yunpan123_openai/"))

	return nil
}

func (y *YunPan123OpenAPIUploader) GetUploadUrl(path string) (string, error) {

	secret := configs.GlobalConfig.Auth.Secret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LocalFileSignJWTClaims{
		Path: path,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		},
	})
	authorization, _ := token.SignedString([]byte(secret))

	return fmt.Sprintf("/uploader/yunpan123_openai/%s?Authorization=%s", path, authorization), nil
}

func (y *YunPan123OpenAPIUploader) GetDownloadUrl(path string) (string, error) {
	if y.signKey == "" {
		return fmt.Sprintf("https://vip.123pan.cn/%s/%s/%s", y.uid, y.dlZone, path), nil
	} else {
		originURL := fmt.Sprintf("https://vip.123pan.cn/%s/%s/%s", y.uid, y.dlZone, path)
		newurl, _ := signURL(originURL, y.signKey, y.uid, time.Second*time.Duration(y.signTime))
		return newurl, nil
	}
}

func (y *YunPan123OpenAPIUploader) Put(savepath string, data []byte) error {
	yunpan123_path := path.Join(y.dlZone, savepath)
	return y.client.writeData(yunpan123_path, data)
}
