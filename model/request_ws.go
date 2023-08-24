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

func (ws WebsocketDetail) Send(debug string, debugMsg *DebugMsg, mongoCollection *mongo.Collection, globalVar *sync.Map) (bool, int64, uint64, float64, float64) {
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

func (ws WebsocketDetail) Request(debug string, debugMsg *DebugMsg, mongoCollection *mongo.Collection, globalVar *sync.Map) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	var conn *websocket.Conn
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	connectionResults, recvResults, writeResults := new(DebugMsg), new(DebugMsg), new(DebugMsg)
	recvResults.Type = "recv"
	recvResults.UUID = debugMsg.UUID
	recvResults.ApiName = debugMsg.ApiName
	recvResults.TeamId = debugMsg.TeamId
	recvResults.ApiId = debugMsg.ApiId
	recvResults.RequestType = debugMsg.RequestType

	writeResults.Type = "send"
	writeResults.UUID = debugMsg.UUID
	writeResults.ApiName = debugMsg.ApiName
	writeResults.TeamId = debugMsg.TeamId
	writeResults.ApiId = debugMsg.ApiId
	writeResults.RequestType = debugMsg.RequestType

	connectionResults.Type = "connection"
	connectionResults.UUID = debugMsg.UUID
	connectionResults.ApiName = debugMsg.ApiName
	connectionResults.TeamId = debugMsg.TeamId
	connectionResults.ApiId = debugMsg.ApiId
	connectionResults.RequestType = debugMsg.RequestType
	headers := map[string][]string{}
	for _, header := range ws.WsHeader {
		if header.IsChecked != constant.Open {
			continue
		}
		headers[header.Var] = []string{header.Val}
	}
	header, _ := json.Marshal(headers)
	if header != nil {
		writeResults.RequestHeader = string(header)
	}
	wsConfig := ws.WsConfig
	ws.Url = strings.TrimSpace(ws.Url)
	recvResults.RequestUrl = ws.Url
	writeResults.RequestUrl = ws.Url
	connectionResults.RequestUrl = ws.Url
	for i := 0; i < wsConfig.RetryNum; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(ws.Url, headers)
		if conn != nil {
			connectionResults.Status = constant.Success
			connectionResults.IsStop = false
			break
		}
	}
	if err != nil || conn == nil {
		if err != nil {
			recvResults.ResponseBody = err.Error()
			writeResults.ResponseBody = err.Error()
			recvResults.Status = constant.Failed
			writeResults.Status = constant.Failed
		} else {
			recvResults.RequestBody = "连接为空"
			writeResults.ResponseBody = "连接为空"
			recvResults.Status = constant.Success
			writeResults.Status = constant.Success
		}

		recvResults.IsStop = true
		writeResults.IsStop = true
		if debug != "stop" {
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
		}

	}

	wsConfig.init()
	readTimeAfter, writeTimeAfter := time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(wsConfig.ConnectDurationTime)*time.Second)
	ticker := time.NewTicker(time.Duration(wsConfig.SendMsgDurationTime) * time.Millisecond)
	defer ticker.Stop()
	switch wsConfig.ConnectType {
	// 长连接
	case constant.LongConnection:
		// 订阅redis中消息  任务状态：包括：报告停止；debug日志状态；任务配置变更
		adjustKey := fmt.Sprintf("WsStatusChange:%s", debugMsg.UUID)
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
								writeResults.RequestBody = err.Error()
							} else {
								writeResults.RequestBody = ""
							}
							writeResults.Status = constant.Failed
							writeResults.IsStop = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults.Status = constant.Failed
						writeResults.IsStop = true
						Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:

						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						if wsStatusChange.Type == 1 {
							writeResults.Status = constant.Failed
							writeResults.IsStop = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}
					case <-ticker.C:
						bodyBytes := []byte(ws.SendMessage)
						err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
						writeResults.RequestBody = ws.SendMessage
						writeResults.Status = constant.Success
						writeResults.IsStop = false
						if err != nil {
							writeResults.RequestBody = err.Error()
							writeResults.Status = constant.Failed
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
								writeResults.RequestBody = err.Error()
							} else {
								writeResults.RequestBody = ""
							}
							writeResults.Status = constant.Failed
							writeResults.IsStop = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults.Status = constant.Success
						writeResults.IsStop = true
						Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:
						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						switch wsStatusChange.Type {
						case 1:
							writeResults.Status = constant.Failed
							writeResults.IsStop = true
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						case 2:
							body := wsStatusChange.Message
							bodyBytes := []byte(body)
							err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
							writeResults.RequestBody = body
							writeResults.Status = constant.Success
							writeResults.IsStop = false
							if err != nil {
								writeResults.RequestBody = err.Error()
								writeResults.Status = constant.Failed
								Insert(mongoCollection, writeResults, middlewares.LocalIp)
								continue
							}
							Insert(mongoCollection, writeResults, middlewares.LocalIp)
						}

					}
				}

			}(wg, pubSub)
		default:
			writeResults.Status = constant.Failed
			writeResults.IsStop = true
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
							recvResults.ResponseBody = err.Error()
						} else {
							recvResults.ResponseBody = ""
						}
						recvResults.Status = constant.Failed
						recvResults.IsStop = true
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				}
				select {
				case <-readTimeAfter:
					recvResults.Status = constant.Success
					recvResults.ResponseBody = ""
					recvResults.IsStop = true
					Insert(mongoCollection, recvResults, middlewares.LocalIp)
					return
				case c := <-statusCh:
					_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
					if wsStatusChange.Type == 1 {
						recvResults.Status = constant.Failed
						recvResults.IsStop = true
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				default:
					m, p, connectErr := conn.ReadMessage()
					if connectErr != nil {
						recvResults.ResponseBody = connectErr.Error()
						recvResults.Status = constant.Failed
						recvResults.IsStop = false
						Insert(mongoCollection, recvResults, middlewares.LocalIp)
						break
					}
					recvResults.Status = constant.Success
					recvResults.ResponseMessageType = m
					recvResults.ResponseBody = string(p)
					recvResults.IsStop = false
					Insert(mongoCollection, recvResults, middlewares.LocalIp)
				}
			}
		}()
		wg.Wait()

	// 短链接
	case constant.ShortConnection:
		if conn == nil {
			if err != nil {
				recvResults.ResponseBody = err.Error()
				writeResults.RequestBody = err.Error()
			} else {
				recvResults.ResponseBody = "连接为空"
				writeResults.RequestBody = "连接为空"
			}
			recvResults.Status = constant.Failed
			writeResults.Status = constant.Failed
			recvResults.IsStop = true
			writeResults.IsStop = true
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}

		bodyBytes := []byte(ws.SendMessage)
		err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
		writeResults.RequestBody = ws.SendMessage
		if err != nil {
			recvResults.ResponseBody = err.Error()
			writeResults.ResponseBody = err.Error()
			recvResults.Status = constant.Failed
			writeResults.Status = constant.Failed
			recvResults.IsStop = true
			writeResults.IsStop = true
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults.IsStop = true
		m, p, err1 := conn.ReadMessage()
		if err1 != nil {
			recvResults.ResponseBody = err1.Error()
			recvResults.IsStop = true
			recvResults.Status = constant.Failed
			Insert(mongoCollection, writeResults, middlewares.LocalIp)
			Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults.Status = constant.Success
		writeResults.Status = constant.Success
		recvResults.ResponseMessageType = m
		recvResults.ResponseBody = string(p)
		recvResults.IsStop = true
		writeResults.IsStop = true
		Insert(mongoCollection, writeResults, middlewares.LocalIp)
		Insert(mongoCollection, recvResults, middlewares.LocalIp)
		return

	default:
		recvResults.Status = constant.Failed
		writeResults.Status = constant.Failed
		recvResults.IsStop = true
		writeResults.IsStop = true
	}
	return
}

func (wsC *WsConfig) init() {
	if wsC.ConnectDurationTime <= 0 {
		wsC.ConnectDurationTime = 1
	}
	if wsC.SendMsgDurationTime <= 0 {
		wsC.SendMsgDurationTime = 1
	}
}
