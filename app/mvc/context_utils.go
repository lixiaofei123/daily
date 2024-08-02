package mvc

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetValueFromAnywhere(ctx echo.Context, key string) string {

	cookie, err := ctx.Cookie(key)
	if err == nil {
		return cookie.Value
	}

	value := ctx.Request().Header.Get(key)
	if value != "" {
		return value
	}

	value = ctx.FormValue(key)
	if value != "" {
		return value
	}

	return ctx.QueryParam(key)

}

func GetIntValueFromAnywhere(ctx echo.Context, key string) int {

	var value string
	cookie, err := ctx.Cookie(key)
	if err == nil {
		value = cookie.Value
	} else {
		value = ctx.Request().Header.Get(key)
		if value == "" {
			value = ctx.FormValue(key)
			if value == "" {
				value = ctx.QueryParam(key)
			}
		}
	}

	if value == "" {
		return 0
	}

	intValue, _ := strconv.Atoi(value)

	return intValue

}

func GetValue(ctx echo.Context, key string) string {
	value := ctx.FormValue(key)
	if value != "" {
		return value
	}
	return ctx.QueryParam(key)
}

func GetParamIntValue(ctx echo.Context, key string) int {
	value := ctx.Param(key)
	numValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return numValue
}

func GetValueWithDefault(ctx echo.Context, key, defaultValue string) string {
	value := ctx.FormValue(key)
	if value != "" {
		return value
	}
	value = ctx.QueryParam(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func GetIntValueWithDefault(ctx echo.Context, key string, defaultValue int) int {
	value := ctx.FormValue(key)
	if value == "" {
		value = ctx.QueryParam(key)
	}
	if value == "" {
		return defaultValue
	}
	numValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return numValue
}

func GetMapValueFromBody(ctx echo.Context) (map[string]interface{}, error) {
	body := ctx.Request().Body
	defer body.Close()
	bodydata, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	resultMap := new(map[string]interface{})
	err = json.Unmarshal(bodydata, resultMap)
	if err != nil {
		return nil, err
	}
	return *resultMap, err

}

func GetBooleanValueWithDefault(ctx echo.Context, key string, defaultValue bool) bool {
	value := ctx.FormValue(key)
	if value == "" {
		value = ctx.QueryParam(key)
	}
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}
