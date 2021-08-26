package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/pfhds/live-stream-ai/middleware"
	"github.com/pfhds/live-stream-ai/models"
	"github.com/pfhds/live-stream-ai/utils/log"
)

type Option func(*gin.Engine)

var options = []Option{}

// 注册app的路由配置
func Include(opts ...Option) {
	options = append(options, opts...)
}

// 初始化
func Init(config *models.Config) *gin.Engine {
	gin.SetMode(config.Server.Mode)
	r := gin.New()
	r.NoRoute(middleware.NotFoundHandle)
	r.NoMethod(middleware.NotFoundHandle)

	r.Use(
		log.NewLocal(config.Server.LogPath),
		log.NewLogConsole(),
		middleware.Recover)
	for _, opt := range options {
		opt(r)
	}

	return r
}
