package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pfhds/live-stream-ai/models"
	"github.com/pfhds/live-stream-ai/utils/convert"
)

type Result struct {
	Ctx *gin.Context
}

const (
	API_STOP     = 403  // 服务器拒绝请求（包括无权限）
	API_NOTFOUND = 404  // 服务器找不到请求的资源
	API_ERROR    = 500  // 服务器内部错误
	API_ERROR_F  = 9999 // API平台错误
)

func NewResult(ctx *gin.Context) *Result {
	return &Result{Ctx: ctx}
}

// 成功
func (r *Result) Success(data interface{}) {
	if data == nil {
		data = gin.H{}
	}

	res := models.ResultJson{}
	res.Code = 0
	res.Data = data
	res.Msg = ""

	r.Ctx.JSON(http.StatusOK, res)
}

// 失败
func (r *Result) Error(code int, msg interface{}) {
	res := models.ResultJson{}
	res.Code = code
	res.Data = nil
	res.Msg = convert.ErrorToString(msg)

	r.Ctx.JSON(http.StatusOK, res)
}

// API接口错误
func (r *Result) ErrorF(msg interface{}) {
	r.Error(API_ERROR_F, msg)
}

// 404
func NotFoundHandle(c *gin.Context) {
	NewResult(c).Error(API_NOTFOUND, "资源未找到")
}

// 500
func Recover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("服务器内部出现错误，原因： %v \n", r)
			err := "服务器内部出现错误，原因：" + convert.ErrorToString(r)
			NewResult(c).Error(API_ERROR, err)

			c.Abort()
		}
	}()

	c.Next()
}
