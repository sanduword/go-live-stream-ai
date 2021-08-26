package main

import (
	"fmt"

	"github.com/pfhds/live-stream-ai/core"
	"github.com/pfhds/live-stream-ai/models"
	"github.com/pfhds/live-stream-ai/routers"
	"github.com/pfhds/live-stream-ai/service"
	"github.com/pfhds/live-stream-ai/utils/conf"
	"github.com/pfhds/live-stream-ai/utils/log"
	"github.com/pfhds/live-stream-ai/utils/redis"
)

func main() {
	log.Init()
	config, err := conf.New()
	if err != nil {
		log.Errorf("Error getting conf configuration file, reason: %v", err)
		return
	}

	redis.Init(&config.Redis)

	if err := run(config); err != nil {
		log.Fatalf("Cannot start server, error: %v", err)
		return
	}

	log.Info("Service started...")
}

func run(conf *models.Config) error {
	// 加载多路由配置
	//routers.Include(router1, router2, ...)
	routers.Include(
		routers.NewStreamController,
		routers.NewHtmlController,
	)

	// 启动service
	service.Init(conf)
	// 启动websocket
	go core.Manager.StartSocket()
	// 启动识别任务
	go core.Manager.StartAiStrem()
	// 启动yolo模型初始化工作
	go core.YoloMs.StartYolo(&conf.Server)

	// 初始化路由
	r := routers.Init(conf)
	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		return err
	}

	return nil
}
