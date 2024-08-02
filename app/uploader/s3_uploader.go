package uploader

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/lixiaofei123/daily/app/utils"
)

func init() {
	RegsiterDriver("S3", &S3Uploader{}, &S3UploaderConfig{})
}

type S3Uploader struct {
	config *aws.Config
	bucket string
	s3     *s3.S3
	domain string
	pkey   string
}

type S3UploaderConfig struct {
	SecretID    string `json:"secretId"`
	SecretKey   string `json:"secretKey"`
	Region      string `json:"region"`
	Endpoint    string `json:"endpoint"`
	Bucket      string `json:"bucket"`
	VirutalHost string `json:"virutalHost"`
	Domain      string `json:"domain"`
	Pkey        string `json:"pkey"`
}

func (u *S3Uploader) initConfig(config interface{}) error {
	s3config := config.(*S3UploaderConfig)

	creds := credentials.NewStaticCredentials(s3config.SecretID, s3config.SecretKey, "")

	virutalHost, _ := strconv.ParseBool(s3config.VirutalHost)
	u.config = &aws.Config{
		Region:           aws.String(s3config.Region),
		Endpoint:         &s3config.Endpoint,
		S3ForcePathStyle: aws.Bool(!virutalHost),
		Credentials:      creds,
	}

	u.bucket = s3config.Bucket
	u.domain = s3config.Domain
	u.pkey = s3config.Pkey

	sess, err := session.NewSession(u.config)
	if err != nil {
		return err
	}
	u.s3 = s3.New(sess)

	return nil

}

func (u *S3Uploader) InitUploader(c *echo.Group) error {
	return nil
}

func (u *S3Uploader) GetUploadUrl(path string) (string, error) {
	req, _ := u.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(path),
	})

	return req.Presign(5 * time.Minute)
}

func (u *S3Uploader) GetDownloadUrl(path string) (string, error) {

	if u.domain != "" {
		if u.pkey == "" {
			return fmt.Sprintf("%s/%s", u.domain, path), nil
		} else {
			// 开启了鉴权
			uri := fmt.Sprintf("/%s", path)
			timestamp := time.Now().Unix()
			needsignstr := fmt.Sprintf("%s%s%d", u.pkey, uri, timestamp)
			sign := utils.Md5(needsignstr)
			return fmt.Sprintf("%s/%s?sign=%s&t=%d", u.domain, path, sign, timestamp), nil
		}

	} else {
		req, _ := u.s3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(u.bucket),
			Key:    aws.String(path),
		})
		return req.Presign(5 * time.Minute)
	}
}

func (u *S3Uploader) Put(path string, data []byte) error {

	_, err := u.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})

	return err
}
