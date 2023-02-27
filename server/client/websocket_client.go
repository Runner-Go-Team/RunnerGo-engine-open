package client

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/gorilla/websocket"
	"time"
)

type WebsocketClient struct {
	Conn               *websocket.Conn
	Addr               *string
	IsAlive            bool
	Timeout            int // 连接超时时间 0 为无限制
	Path               string
	SendMsgChan        chan string
	RecvMsgChan        chan string
	MaxConnection      int // 最大连接数
	ConnectionDuration int // 重新连接间隔
	MaxContent         int // 允许的最大发送内容大小， 0 为无限制
}

func WebSocketRequest(url string, body string, headers map[string][]string, timeout int) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	websocketClient := NewWsClientManager(url, timeout)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, connecting to : %s", middlewares.LocalIp, url))
	if websocketClient.IsAlive == false {
		for i := 0; i < 3; i++ {
			startTime := time.Now().UnixMilli()
			websocketClient.Conn, _, err = websocket.DefaultDialer.Dial(url, headers)
			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				log.Logger.Error(fmt.Sprintf("机器ip:%s,  第 %d 次connecting to: %s 失败!", middlewares.LocalIp, i, url))
				continue
			}
			websocketClient.IsAlive = true
			bodyBytes := []byte(body)
			err = websocketClient.Conn.WriteMessage(websocket.TextMessage, bodyBytes)
			sendBytes = uint(len(body))
			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				log.Logger.Error(fmt.Sprintf("机器ip:%s, 第 %d 次向: %s写消息失败失败!", middlewares.LocalIp, i, url))
				continue
			}

			_, resp, err = websocketClient.Conn.ReadMessage()

			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				websocketClient.IsAlive = false
				// 出现错误，退出读取，尝试重连
				continue
			}
			//requestTime = tools.TimeDifference(startTime)
			requestTime = tools.TimeDifference(startTime)
			break
		}
	}
	return

}

// NewWsClientManager 构造函数
func NewWsClientManager(url string, timeout int) *WebsocketClient {
	var conn *websocket.Conn
	return &WebsocketClient{
		Addr:    &url,
		Conn:    conn,
		IsAlive: false,
		Timeout: timeout,
	}
}
