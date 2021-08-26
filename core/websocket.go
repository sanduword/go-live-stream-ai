package core

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pfhds/live-stream-ai/utils/log"
)

// 服务端管理
var Manager = ClientManager{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[string]*Client),
}

func createId(index string) string {
	return "live_stream_" + index
}

// 返回客户端消息
func resMessage(index string, status int, content interface{}) (message []byte) {
	messageStruct := &Message{}
	messageStruct.Index = index
	messageStruct.Status = status
	messageStruct.Content = content
	client := Manager.Clients[createId(index)]
	if (client != nil && client != &Client{}) {
		messageStruct.StartTime = client.StartTime
		messageStruct.CurTime = client.CurTime
		if messageStruct.CurTime.IsZero() {
			messageStruct.CurTime = time.Now()
		}
	}

	json, _ := json.Marshal(messageStruct)
	return json
}

// 启动一个socket协程任务
func (manager *ClientManager) StartSocket() {
	for {
		log.Info("信号管道通信...")
		select {
		case conn := <-Manager.Register:
			log.Infof("进来一个新识别请求 %v", conn.ID)
			Manager.Clients[conn.ID] = conn
			conn.Send <- resMessage(conn.Index, NO, "连接Socket服务端成功")
		case conn := <-Manager.Unregister:
			log.Infof("识别请求已退出 %v", conn.ID)
			if _, ok := Manager.Clients[conn.ID]; ok {
				conn.Send <- resMessage(conn.Index, ERROR, "有一个识别请求已断开连接")
				conn.Status = ERROR
				conn.execAiTask()
				close(conn.Send)
				//close(conn.BaseImg)
				delete(Manager.Clients, conn.ID)
			}
		case message := <-Manager.Broadcast:
			messageStruct := Message{}
			json.Unmarshal(message, &messageStruct)
			for id, conn := range Manager.Clients {
				if id != createId(messageStruct.Index) {
					continue
				}

				select {
				case conn.Send <- resMessage(conn.Index, conn.Status, messageStruct.Content):
				default:
					close(conn.Send)
					//close(conn.BaseImg)
					delete(Manager.Clients, conn.ID)
				}
			}
		}
	}
}

// 启动一个识别的协程
func (manager *ClientManager) StartAiStrem() {
	for {
		for _, conn := range Manager.Clients {
			switch conn.Status {
			case NO:
				conn.execAiTask()
				conn.Status = NOW
				conn.StartTime = time.Now()
			case NOW:
				conn.CurTime = time.Now()
			case ERROR:
				conn.execAiTask()
			}

		}
		time.Sleep(time.Second * 1)
	}
}

// 处理AI识别任务
func (c *Client) execAiTask() {
	switch c.Status {
	case NO:
		log.Infoln("启动识别任务")
		go c.StartOpencvStream()
	case ERROR:
		log.Infoln("停止识别任务")
	}
}

// 读取客户端发送过来的信息
func (c *Client) ReadSocket() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
	}()

	for {
		c.Socket.PongHandler()
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- c
			c.Socket.Close()
			break
		}

		log.Infof("接收到客户端信息：%v", string(message[:]))
		Manager.Broadcast <- message
	}
}

// 往客户端推送数据
func (c *Client) WriteSocket() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.TextMessage, resMessage(c.Index, ERROR, "系统异常"))
				return
			}
			log.Infof("发送给客户端的消息: %v", string(message))
			c.Socket.WriteMessage(websocket.TextMessage, message)
		case baseimg, ok := <-c.BaseImg:
			if !ok {
				log.Error("接收视频帧异常")
				return
			}
			c.Socket.WriteMessage(websocket.BinaryMessage, baseimg)
		case warndata, ok := <-c.WarnData:
			if !ok {
				log.Error("接收识别结果异常")
				return
			}
			c.Socket.WriteMessage(websocket.TextMessage, resMessage(c.Index, NOW, warndata))
		}
	}
}

// socker 协议、用户认证、自定义信息
func WsHandler(index string, url string, writer gin.ResponseWriter, request *http.Request) {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}).Upgrade(writer, request, nil)
	if err != nil {
		http.NotFound(writer, request)
		return
	}

	client := &Client{
		ID:       createId(index),
		Index:    index,
		Url:      url,
		Socket:   conn,
		Send:     make(chan []byte),
		BaseImg:  make(chan []byte),
		WarnData: make(chan []*LabelResult),
	}

	Manager.Register <- client

	go client.ReadSocket()
	go client.WriteSocket()
}

// 获取最大数量
func GetMaxWs() int {
	return len(Manager.Clients)
}
