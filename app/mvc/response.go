package mvc

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RespMessage struct {
	Code     int
	Text     string
	Data     interface{}
	Err      error
	Template string
}

type PageRespMessage struct {
	PageIndex  int         `json:"pageIndex"`
	PageCount  int         `json:"pageCount"`
	TotalCount int         `json:"totalCount"`
	List       interface{} `json:"list"`
}

type TextResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DataResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type EmptyResponse struct {
	Code int `json:"code"`
}

// implementsResult.
func (e RespMessage) Dispatch(ctx echo.Context) {
	if e.Code == 0 {
		e.Code = 200
	}

	if ctx.Response().Status == http.StatusNotModified {
		ctx.JSON(http.StatusNotModified, EmptyResponse{
			Code: e.Code,
		})
		return
	}

	if ctx.Response().Status == http.StatusMovedPermanently {
		ctx.JSON(http.StatusMovedPermanently, EmptyResponse{
			Code: e.Code,
		})
		return
	}

	if ctx.Response().Status == http.StatusFound {
		ctx.JSON(http.StatusFound, EmptyResponse{
			Code: e.Code,
		})
		return
	}

	if e.Template != "" {
		err := ctx.Render(e.Code, e.Template, e.Data)
		if err != nil {
			log.Println(err.Error())
		}
	} else if e.Text != "" {
		ctx.JSON(e.Code, TextResponse{
			Code:    e.Code,
			Message: e.Text,
		})
	} else if e.Data != nil {
		ctx.JSON(e.Code, DataResponse{
			Code: e.Code,
			Data: e.Data,
		})
	} else if e.Err.Error() != "" {
		ctx.JSON(e.Code, DataResponse{
			Code: e.Code,
			Data: e.Err.Error(),
		})
	} else {
		ctx.JSON(e.Code, EmptyResponse{
			Code: e.Code,
		})
	}
}

func Error(code int, err error) Result {
	return RespMessage{
		Code: code,
		Err:  err,
	}
}

func Ok(data interface{}) Result {

	return RespMessage{
		Code: 200,
		Data: data,
	}
}

func OkHtml(template string, data interface{}) Result {
	return RespMessage{
		Code:     200,
		Template: template,
		Data:     data,
	}
}

func ErrorHtml(code int, err error) Result {
	return RespMessage{
		Code:     200,
		Template: "error.html",
		Data:     err.Error(),
	}
}
