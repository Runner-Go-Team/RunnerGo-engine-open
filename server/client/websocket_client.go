package client

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

func WebSocketRequest(url string, body string, headers map[string][]string, wsConfig model.WsConfig) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	var conn *websocket.Conn
	conn, _, err = websocket.DefaultDialer.Dial(url, headers)
	if err != nil || conn == nil {
		for i := 0; i < wsConfig.RetryNum; i++ {
			conn, _, _ = websocket.DefaultDialer.Dial(url, headers)
			if conn != nil {
				break
			}
		}
	}
	readTimeAfter, writeTimeAfter := time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second)
	ticker := time.NewTicker(time.Duration(wsConfig.SendMsgDurationTime) * time.Millisecond)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, connecting to : %s", middlewares.LocalIp, url))
	switch wsConfig.ConnectType {
	case 1:
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go write(conn, wg, url, body, headers, wsConfig, writeTimeAfter, ticker)
		wg.Add(1)
		go read(conn, wg, url, headers, wsConfig, readTimeAfter)
	case 2:
		if conn == nil {
			return
		}

		bodyBytes := []byte(body)
		err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
		//sendBytes := uint(len(body))
		if err != nil {
			//requestTime = tools.TimeDifference(startTime)
			log.Logger.Error(fmt.Sprintf("ws发送消息失败：%s ", err.Error()))
		}

		m, p, err := conn.ReadMessage()
		if err != nil {
			log.Logger.Debug("ws读取消息错误：", err.Error())
		}
		log.Logger.Debug(fmt.Sprintf("ws消息类型：%d,   读取到的消息：%s     ", m, string(p)))

	}
	return

}

func write(conn *websocket.Conn, wg *sync.WaitGroup, url string, body string, headers map[string][]string, wsConfig model.WsConfig, timeAfter <-chan time.Time, ticker *time.Ticker) {
	defer wg.Done()
	if conn == nil {
		conn, _, _ = websocket.DefaultDialer.Dial(url, headers)
		if conn == nil {
			for i := 0; i < wsConfig.RetryNum; i++ {
				conn, _, _ = websocket.DefaultDialer.Dial(url, headers)
				if conn != nil {
					break
				}
			}
		}
	}
	for {
		if conn == nil {
			return
		}
		select {
		case <-timeAfter:
			return
		case <-ticker.C:
			bodyBytes := []byte(body)
			err := conn.WriteMessage(websocket.TextMessage, bodyBytes)
			//sendBytes := uint(len(body))
			if err != nil {
				//requestTime = tools.TimeDifference(startTime)
				log.Logger.Error(fmt.Sprintf("ws发送消息失败： %s", err.Error()))
				continue
			}
		}
	}

}

func read(conn *websocket.Conn, wg *sync.WaitGroup, url string, headers map[string][]string, wsConfig model.WsConfig, timeAfter <-chan time.Time) {
	defer wg.Done()
	if conn == nil {
		conn, _, _ = websocket.DefaultDialer.Dial(url, headers)
		if conn == nil {
			for i := 0; i < wsConfig.RetryNum; i++ {
				conn, _, _ = websocket.DefaultDialer.Dial(url, headers)
				if conn != nil {
					break
				}
			}
		}
	}
	for {
		if conn == nil {
			return
		}
		select {
		case <-timeAfter:
			return
		default:
			m, p, err := conn.ReadMessage()
			if err != nil {
				log.Logger.Debug("ws读取消息错误：", err.Error())
				break
			}
			log.Logger.Debug(fmt.Sprintf("ws消息类型：%d,   读取到的消息：%s     ", m, string(p)))
		}
	}
}
