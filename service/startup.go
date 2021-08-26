package service

import (
	"github.com/pfhds/live-stream-ai/models"
)

type Dao struct {
	config *models.Config
}

var dao *Dao

func Init(config *models.Config) {
	dao = &Dao{
		config: config,
	}
}
