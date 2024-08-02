package controller

import (
	"strings"

	"github.com/golang-jwt/jwt"
	echo "github.com/labstack/echo/v4"
)

func GetIntValueFromMap(bodyMap map[string]interface{}, key string) *int {
	if value, ok := bodyMap[key]; ok {
		if value0, ok := value.(int); ok {
			return &value0
		}
	}
	return nil
}

func GetStringValueFromMap(bodyMap map[string]interface{}, key string) *string {
	if value, ok := bodyMap[key]; ok {
		if value0, ok := value.(string); ok {
			return &value0
		}
	}
	return nil
}

func GetBoolValueFromMap(bodyMap map[string]interface{}, key string) *bool {
	if value, ok := bodyMap[key]; ok {
		if value0, ok := value.(bool); ok {
			return &value0
		}
	}
	return nil
}

type UserJWTClaims struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Enable   bool   `json:"enable"`
	UserId   uint   `json:"uid"`
	jwt.StandardClaims
}

func IsMobileUserAgent(ctx echo.Context) bool {
	userAgent := ctx.Request().Header.Get("User-Agent")

	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPod", "iPad", "Pad", "BlackBerry", "Opera Mini", "IEMobile", "WPDesktop",
	}
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}
