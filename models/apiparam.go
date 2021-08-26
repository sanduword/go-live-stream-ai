package models

import "time"

const (
	No    = 0
	Run   = 1
	Errpr = 2
)

type InputModel struct {
	// binding:"required"修饰的字段，若接收为空值，则报错，是必须字段
	Url   string `form:"url" json:"url" uri:"url" xml:"url" binding:"required"`
	Index string `form:"index" json:"index" uri:"index" xml:"index" binding:"required"`
}

type OutModel struct {
	HttpLive   string `json:"httplive" binding:"required"`
	HttpResult string `json:"httpresult" binding:"required"`
}

type StreamModel struct {
	InputModel
	// 0 未运行 1 正在运行 2 运行失败
	Status    int       `json:"status" binding:"required"`
	StartTime time.Time `json:"starttime" binding:"required"`
	CurTime   time.Time `json:"curtime" binding:"required"`
	HttpLive  string    `json:"httplive" binding:"required"`
}
