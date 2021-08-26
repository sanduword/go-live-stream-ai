package models

// 统一返回JSON结构体
type ResultJson struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
