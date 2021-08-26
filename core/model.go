package core

import (
	"time"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

const (
	NO    = 0 // 未运行
	NOW   = 1 // 正在运行
	ERROR = 2 // 运行失败
)

// 客户端管理
type ClientManager struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// 客户端 socket
type Client struct {
	ID        string
	Index     string
	Url       string
	Status    int       // 0 未运行 1 正在运行 2 运行失败
	StartTime time.Time `json:"starttime,omitempty"`
	CurTime   time.Time `json:"curtime,omitempty"`
	Socket    *websocket.Conn
	Send      chan []byte
	OpenCv    *gocv.VideoCapture
	BaseImg   chan []byte
	WarnData  chan []*LabelResult
}

// 消息内容
type Message struct {
	Index     string      `json:"index,omitempty"`
	Url       string      `json:"url,omitempty"`
	Content   interface{} `json:"content,omitempty"`
	Status    int         `json:"status"` // 0 未运行 1 正在运行 2 运行失败
	StartTime time.Time   `json:"starttime,omitempty"`
	CurTime   time.Time   `json:"curtime,omitempty"`
}

// Yolo模型管理
type YoloManager struct {
	Labels         []string
	Net            *gocv.Net
	OutputNames    []string
	ScoreThreshold float32
	NmsThreshold   float32
	LabelData      map[string]*LabelResult
}

// 识别结果汇总
type LabelResult struct {
	LabelName string `json:"label"`
	Warn      int    `json:"warn"`
}
