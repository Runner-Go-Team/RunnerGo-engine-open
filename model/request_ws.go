package model

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"sync"
	"time"
)

type WebsocketDetail struct {
	Url            string          `json:"url"`
	Debug          string          `json:"debug"`
	SendMessage    string          `json:"send_message"`
	MessageType    string          `json:"message_type"` // "Binary"、"Text"、"Json"、"Xml"
	WsHeader       []WsQuery       `json:"ws_header"`
	WsParam        []WsQuery       `json:"ws_param"`
	WsEvent        []WsQuery       `json:"ws_event"`
	WsConfig       WsConfig        `json:"ws_config"`
	Configuration  *Configuration  `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	WsVariable     *GlobalVariable `json:"api_variable"`
}

type WsConfig struct {
	ConnectType         int32 `json:"connect_type"`           // 连接类型：1-长连接，2-短连接
	IsAutoSend          int32 `json:"is_auto_send"`           // 是否自动发送消息：0-非自动，1-自动
	ConnectDurationTime int   `json:"connect_duration_time"`  // 连接持续时长，单位：秒
	SendMsgDurationTime int   `json:"send_msg_duration_time"` // 发送消息间隔时长，单位：毫秒
	ConnectTimeoutTime  int   `json:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int   `json:"retry_num"`              // 重连次数
	RetryInterval       int   `json:"retry_interval"`         // 重连间隔时间，单位：毫秒
}

type WsQuery struct {
	IsChecked int32  `json:"is_checked"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}

func (ws WebsocketDetail) Send(debug string, debugMsg map[string]interface{}, mongoCollection *mongo.Collection, globalVar *sync.Map) (bool, int64, uint64, float64, float64) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = constant.NoError
		receivedBytes = float64(0)
	)

	resp, requestTime, sendBytes, err := ws.Request(debug, debugMsg, mongoCollection, globalVar)

	if err != nil {
		isSucceed = false
		errCode = constant.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		receivedBytes, _ = decimal.NewFromFloat(float64(len(resp)) / 1024).Round(2).Float64()
	}
	return isSucceed, errCode, requestTime, float64(sendBytes), receivedBytes
}

func (ws WebsocketDetail) Request(debug string, debugMsg map[string]interface{}, mongoCollection *mongo.Collection, globalVar *sync.Map) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	var conn *websocket.Conn
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	connectionResults, recvResults, writeResults := make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{})
	recvResults["type"] = "recv"
	recvResults["uuid"] = debugMsg["uuid"]
	recvResults["name"] = debugMsg["name"]
	recvResults["team_id"] = debugMsg["team_id"]
	recvResults["target_id"] = debugMsg["target_id"]
	recvResults["request_type"] = debugMsg["request_type"]

	writeResults["type"] = "send"
	writeResults["uuid"] = debugMsg["uuid"]
	writeResults["name"] = debugMsg["name"]
	writeResults["team_id"] = debugMsg["team_id"]
	writeResults["target_id"] = debugMsg["target_id"]
	writeResults["request_type"] = debugMsg["request_type"]

	connectionResults["type"] = "connection"
	connectionResults["uuid"] = debugMsg["uuid"]
	connectionResults["name"] = debugMsg["name"]
	connectionResults["team_id"] = debugMsg["team_id"]
	connectionResults["target_id"] = debugMsg["target_id"]
	connectionResults["request_type"] = debugMsg["request_type"]
	headers := map[string][]string{}
	for _, header := range ws.WsHeader {
		if header.IsChecked != constant.Open {
			continue
		}
		headers[header.Var] = []string{header.Val}
	}
	header, _ := json.Marshal(headers)
	if header != nil {
		writeResults["request_header"] = string(header)
	} else {
		writeResults["request_header"] = ""
	}
	recvResults["type"] = "recv"
	wsConfig := ws.WsConfig
	ws.Url = strings.TrimSpace(ws.Url)
	for i := 0; i < wsConfig.RetryNum; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(ws.Url, headers)
		if conn != nil {
			connectionResults["status"] = true
			connectionResults["is_stop"] = false
			break
		}
	}
	if err != nil || conn == nil {
		if err != nil {
			recvResults["response_body"] = err.Error()
			writeResults["response_body"] = err.Error()
			recvResults["status"] = false
			writeResults["status"] = false
		} else {
			recvResults["request_body"] = "连接为空"
			writeResults["response_body"] = "连接为空"
			recvResults["status"] = true
			writeResults["status"] = true
		}

		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		if debug != "stop" {
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
		}

	}

	if wsConfig.ConnectDurationTime == 0 {
		wsConfig.ConnectDurationTime = 1
	}
	if wsConfig.SendMsgDurationTime == 0 {
		wsConfig.SendMsgDurationTime = 1
	}
	readTimeAfter, writeTimeAfter := time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second)
	ticker := time.NewTicker(time.Duration(wsConfig.SendMsgDurationTime) * time.Millisecond)
	defer ticker.Stop()
	switch wsConfig.ConnectType {
	// 长连接
	case constant.LongConnection:
		// 订阅redis中消息  任务状态：包括：报告停止；debug日志状态；任务配置变更
		adjustKey := fmt.Sprintf("WsStatusChange:%s", debugMsg["uuid"])
		pubSub := SubscribeMsg(adjustKey)
		statusCh := pubSub.Channel()
		var wsStatusChange = new(ConnectionStatusChange)
		wg := new(sync.WaitGroup)
		switch wsConfig.IsAutoSend {
		// 自动发送
		case constant.AutoConnectionSend:
			wg.Add(1)
			go func(wsWg *sync.WaitGroup, sub *redis.PubSub) {
				defer wsWg.Done()
				for {
					if conn == nil {
						for i := 0; i < wsConfig.RetryNum; i++ {
							conn, _, err = websocket.DefaultDialer.Dial(ws.Url, headers)
							if conn != nil {
								break
							}
						}
						if conn == nil {
							if err != nil {
								writeResults["request_body"] = err.Error()
							} else {
								writeResults["request_body"] = ""
							}
							writeResults["status"] = false
							writeResults["is_stop"] = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults["status"] = false
						writeResults["is_stop"] = true
						Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:

						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						if wsStatusChange.Type == 1 {
							writeResults["status"] = false
							writeResults["is_stop"] = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}
					case <-ticker.C:
						bodyBytes := []byte(ws.SendMessage)
						err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
						writeResults["request_body"] = ws.SendMessage
						writeResults["status"] = true
						writeResults["is_stop"] = false
						if err != nil {
							writeResults["request_body"] = err.Error()
							writeResults["status"] = false
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							continue
						}
						Insert(mongoCollection, writeResults, middlewares.LocalIp)
					}
				}

			}(wg, pubSub)
		// 手动发送
		case constant.ConnectionAndSend:
			wg.Add(1)
			go func(wsWg *sync.WaitGroup, sub *redis.PubSub) {
				defer wsWg.Done()
				for {

					if conn == nil {
						for i := 0; i < wsConfig.RetryNum; i++ {
							conn, _, err = websocket.DefaultDialer.Dial(ws.Url, headers)
							if conn != nil {
								break
							}
						}
						if conn == nil {
							if err != nil {
								writeResults["request_body"] = err.Error()
							} else {
								writeResults["request_body"] = ""
							}
							writeResults["status"] = false
							writeResults["is_stop"] = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults["status"] = false
						writeResults["is_stop"] = true
						Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:
						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						switch wsStatusChange.Type {
						case 1:
							writeResults["status"] = false
							writeResults["is_stop"] = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						case 2:
							body := wsStatusChange.Message
							bodyBytes := []byte(body)
							err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
							writeResults["request_body"] = body
							writeResults["status"] = true
							writeResults["is_stop"] = false
							if err != nil {
								writeResults["request_body"] = err.Error()
								writeResults["status"] = false
								Insert(mongoCollection, writeResults, middlewares.LocalIp)
								continue
							}
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
						}

					}
				}

			}(wg, pubSub)
		default:
			writeResults["status"] = false
			writeResults["is_stop"] = true
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			return
		}

		// 读消息
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {

				if conn == nil {
					for i := 0; i < wsConfig.RetryNum; i++ {
						conn, _, err = websocket.DefaultDialer.Dial(ws.Url, headers)
						if conn != nil {
							break
						}
					}
					if conn == nil {
						if err != nil {
							recvResults["response_body"] = err.Error()
						} else {
							recvResults["response_body"] = ""
						}
						recvResults["status"] = false
						recvResults["is_stop"] = true
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				}
				select {
				case <-readTimeAfter:
					recvResults["status"] = true
					recvResults["response_body"] = ""
					recvResults["is_stop"] = true
					Insert(mongoCollection, recvResults, middlewares.LocalIp)
					return
				case c := <-statusCh:
					_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
					if wsStatusChange.Type == 1 {
						recvResults["status"] = false
						recvResults["is_stop"] = true
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				default:
					m, p, connectErr := conn.ReadMessage()
					if connectErr != nil {
						recvResults["response_body"] = connectErr.Error()
						recvResults["status"] = false
						recvResults["is_stop"] = false
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						break
					}
					recvResults["status"] = true
					recvResults["response_message_type"] = m
					recvResults["response_body"] = string(p)
					recvResults["is_stop"] = false
					Insert(mongoCollection, recvResults, middlewares.LocalIp)
				}
			}
		}()
		wg.Wait()

	// 短链接
	case constant.ShortConnection:
		if conn == nil {
			if err != nil {
				recvResults["response_body"] = err.Error()
				writeResults["request_body"] = err.Error()
			} else {
				recvResults["response_body"] = "连接为空"
				writeResults["request_body"] = "连接为空"
			}
			recvResults["status"] = false
			writeResults["status"] = false
			recvResults["is_stop"] = true
			writeResults["is_stop"] = true
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}

		bodyBytes := []byte(ws.SendMessage)
		err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
		writeResults["request_body"] = ws.SendMessage
		if err != nil {
			recvResults["response_body"] = err.Error()
			writeResults["request_body"] = err.Error()
			recvResults["status"] = false
			writeResults["status"] = false
			recvResults["is_stop"] = true
			writeResults["is_stop"] = true
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults["is_stop"] = true
		m, p, err1 := conn.ReadMessage()
		if err1 != nil {
			recvResults["response_body"] = err1.Error()
			recvResults["is_stop"] = true
			recvResults["status"] = false
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults["status"] = true
		writeResults["status"] = true
		recvResults["response_message_type"] = m
		recvResults["response_body"] = string(p)
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		Insert(mongoCollection, writeResults, middlewares.LocalIp)
		Insert(mongoCollection, recvResults, middlewares.LocalIp)
		return

	default:
		recvResults["status"] = false
		writeResults["status"] = false
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
	}
	return
}
