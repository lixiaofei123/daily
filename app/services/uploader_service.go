package services

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/uploader"
	"github.com/lixiaofei123/daily/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"gopkg.in/natefinch/lumberjack.v2"
)

type UploadService interface {
	GetUploadUrl(ctx echo.Context, path string) (string, error)

	GetDownloadUrl(ctx echo.Context, path string) (string, error)

	PutData(path string, data []byte) error
}

func NewUploadService(c *echo.Group) UploadService {

	uploader0 := uploader.NewUploader(configs.GlobalConfig.Uploader.Name, configs.GlobalConfig.Uploader.Config)
	err := uploader0.InitUploader(c)
	if err != nil {
		log.Panic(err)
	}

	ratelimit := configs.GlobalConfig.Uploader.RateLimit
	var limiters []*rate.Limiter = nil
	if ratelimit != nil {
		log.Println("设置了限流")
		ratelimitRoles := configs.GlobalConfig.Uploader.RateLimit.Roles
		ratelimitRoleArr := strings.Split(ratelimitRoles, ";")
		limiters = []*rate.Limiter{}
		if len(ratelimitRoleArr) > 0 {
			for _, ratelimitRole := range ratelimitRoleArr {
				arr := strings.Split(ratelimitRole, ":")
				if len(arr) == 2 {
					tokens, err := strconv.Atoi(arr[0])
					if err == nil {
						seconds, err := strconv.Atoi(arr[1])
						if err == nil {

							limiter := rate.NewLimiter(rate.Every(time.Duration(seconds)*time.Second/time.Duration(tokens)), tokens)
							limiters = append(limiters, limiter)
						}
					}
				}
			}
		}

	} else {
		log.Println("未设置限流，请注意流量访问")
	}

	loggerConf := configs.GlobalConfig.Uploader.Logger

	var logger *zap.SugaredLogger = nil
	if loggerConf != nil && loggerConf.Path != "" {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   loggerConf.Path + "/access.log",
				MaxSize:    500,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   true,
			}),
			zap.NewAtomicLevelAt(zapcore.InfoLevel),
		)

		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)).Sugar()
	}

	return &uploaderService{
		uploader: uploader0,
		limiters: limiters,
		logger:   logger,
	}
}

type uploaderService struct {
	uploader uploader.Uploader
	limiters []*rate.Limiter
	logger   *zap.SugaredLogger
}

func (u *uploaderService) GetUploadUrl(ctx echo.Context, path string) (string, error) {
	return u.uploader.GetUploadUrl(path)
}

func (u *uploaderService) GetDownloadUrl(ctx echo.Context, path string) (string, error) {

	allow := true
	if len(u.limiters) > 0 {
		for _, limiter := range u.limiters {
			if !limiter.Allow() {
				allow = false
				break
			}
		}
	}

	if u.logger != nil {
		u.logger.Infow("access log",
			"ip", ctx.RealIP(),
			"method", ctx.Request().Method,
			"driver", configs.GlobalConfig.Uploader.Name,
			"path", path,
			"ua", ctx.Request().Header.Get("User-Agent"),
			"referer", ctx.Request().Header.Get("Referer"),
			"isAllow", allow,
		)
	}

	if allow {
		return u.uploader.GetDownloadUrl(path)
	} else {
		return "", apperr.ErrTooManyRequests
	}

}

func (u *uploaderService) PutData(path string, data []byte) error {
	return u.uploader.Put(path, data)
}
