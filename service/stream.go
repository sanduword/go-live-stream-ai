package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pfhds/live-stream-ai/core"
	"github.com/pfhds/live-stream-ai/middleware"
)

// 处理实时流
func ExecRealStream(c *gin.Context) {
	result := middleware.NewResult(c)
	if core.GetMaxWs() > dao.config.Server.MaxAi {
		result.ErrorF(fmt.Sprintf("A maximum of %d live streams can be processed", dao.config.Server.MaxAi))
		return
	}
	index := c.Query("index")
	url := c.Query("url")
	if len(index) == 0 || len(url) == 0 {
		result.ErrorF("The live broadcast ID and address cannot be empty")
		return
	}

	core.WsHandler(index, url, c.Writer, c.Request)
}

func Pong(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
