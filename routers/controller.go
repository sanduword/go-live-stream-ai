package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/pfhds/live-stream-ai/service"
)

// 处理实时流相关路由
func NewStreamController(e *gin.Engine) {
	e.GET("/execstream", service.ExecRealStream)
	e.GET("/pong", service.Pong)
}

// HTML模板
func NewHtmlController(e *gin.Engine) {
	e.LoadHTMLGlob("web/*")
	e.GET("/ai", service.ExecWebHtml)
}
