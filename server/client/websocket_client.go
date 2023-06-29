package client

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

func WebSocketRequest(recvResults, writeResults, connectionResults map[string]interface{}, mongoCollection *mongo.Collection, url string, body string, headers map[string][]string, wsConfig model.WsConfig, uid uuid.UUID) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	var conn *websocket.Conn
	recvResults["type"] = "recv"
	for i := 0; i < wsConfig.RetryNum; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, headers)
		if conn != nil {
			connectionResults["status"] = constant.Success
			connectionResults["is_stop"] = false
			break
		}
	}
	if err != nil || conn == nil {
		if err != nil {
			recvResults["err"] = err.Error()
			writeResults["err"] = err.Error()
			recvResults["status"] = constant.Failed
			writeResults["status"] = constant.Failed
		} else {
			recvResults["err"] = ""
			writeResults["err"] = ""
			recvResults["status"] = constant.Success
			writeResults["status"] = constant.Success
		}

		recvResults["request_body"] = ""
		writeResults["response_body"] = ""
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
		model.Insert(mongoCollection, recvResults, middlewares.LocalIp)

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
		adjustKey := fmt.Sprintf("WsStatusChange:%s", uid.String())
		pubSub := model.SubscribeMsg(adjustKey)
		statusCh := pubSub.Channel()
		var wsStatusChange = new(model.ConnectionStatusChange)
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
							conn, _, err = websocket.DefaultDialer.Dial(url, headers)
							if conn != nil {
								break
							}
						}
						if conn == nil {
							if err != nil {
								writeResults["err"] = err.Error()
							} else {
								writeResults["err"] = ""
							}
							writeResults["status"] = constant.Failed
							writeResults["is_stop"] = true
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults["status"] = constant.Failed
						writeResults["is_stop"] = true
						model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:

						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						if wsStatusChange.Type == 1 {
							writeResults["status"] = constant.Failed
							writeResults["is_stop"] = true
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}
					case <-ticker.C:
						bodyBytes := []byte(body)
						err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
						writeResults["request_body"] = body
						writeResults["status"] = constant.Success
						writeResults["is_stop"] = false
						if err != nil {
							writeResults["err"] = err.Error()
							writeResults["status"] = constant.Failed
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
							continue
						}
						model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
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
							conn, _, err = websocket.DefaultDialer.Dial(url, headers)
							if conn != nil {
								break
							}
						}
						if conn == nil {
							if err != nil {
								writeResults["err"] = err.Error()
							} else {
								writeResults["err"] = ""
							}
							writeResults["status"] = constant.Failed
							writeResults["is_stop"] = true
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						}

					}
					select {
					case <-writeTimeAfter:
						writeResults["status"] = constant.Failed
						writeResults["is_stop"] = true
						model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					case c := <-statusCh:
						_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
						switch wsStatusChange.Type {
						case 1:
							writeResults["status"] = constant.Failed
							writeResults["is_stop"] = true
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
							return
						case 2:
							body = wsStatusChange.Message
							bodyBytes := []byte(body)
							err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
							writeResults["request_body"] = body
							writeResults["status"] = constant.Success
							writeResults["is_stop"] = false
							if err != nil {
								writeResults["err"] = err.Error()
								writeResults["status"] = constant.Failed
								model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
								continue
							}
							model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
						}

					}
				}

			}(wg, pubSub)
		default:
			writeResults["status"] = constant.Failed
			writeResults["is_stop"] = true
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			return
		}

		// 读消息
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {

				if conn == nil {
					for i := 0; i < wsConfig.RetryNum; i++ {
						conn, _, err = websocket.DefaultDialer.Dial(url, headers)
						if conn != nil {
							break
						}
					}
					if conn == nil {
						if err != nil {
							recvResults["err"] = err.Error()
						} else {
							recvResults["err"] = ""
						}
						recvResults["status"] = constant.Failed
						recvResults["is_stop"] = true
						model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				}
				select {
				case <-readTimeAfter:
					recvResults["status"] = constant.Success
					recvResults["err"] = ""
					recvResults["is_stop"] = true
					model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
					return
				case c := <-statusCh:
					_ = json.Unmarshal([]byte(c.Payload), wsStatusChange)
					if wsStatusChange.Type == 1 {
						recvResults["status"] = constant.Failed
						recvResults["is_stop"] = true
						model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				default:
					m, p, connectErr := conn.ReadMessage()
					if connectErr != nil {
						recvResults["err"] = connectErr.Error()
						recvResults["status"] = constant.Failed
						recvResults["is_stop"] = false
						model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
						break
					}
					recvResults["status"] = constant.Success
					recvResults["err"] = ""
					recvResults["response_message_type"] = m
					recvResults["response_body"] = string(p)
					recvResults["is_stop"] = false
					model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
				}
			}
		}()
		wg.Wait()

	// 短链接
	case constant.ShortConnection:
		if conn == nil {
			if err != nil {
				recvResults["err"] = err.Error()
				writeResults["err"] = err.Error()
			} else {
				recvResults["err"] = ""
				writeResults["err"] = ""
			}
			recvResults["status"] = constant.Failed
			writeResults["status"] = constant.Failed
			recvResults["is_stop"] = true
			writeResults["is_stop"] = true
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}

		bodyBytes := []byte(body)
		err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
		writeResults["request_body"] = body
		if err != nil {
			recvResults["err"] = err.Error()
			writeResults["err"] = err.Error()
			recvResults["status"] = constant.Failed
			writeResults["status"] = constant.Failed
			recvResults["is_stop"] = true
			writeResults["is_stop"] = true
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults["is_stop"] = true
		m, p, err1 := conn.ReadMessage()
		if err1 != nil {
			recvResults["err"] = err1.Error()
			recvResults["is_stop"] = true
			recvResults["status"] = constant.Failed
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults["status"] = constant.Success
		writeResults["status"] = constant.Success
		recvResults["response_message_type"] = m
		recvResults["response_body"] = string(p)
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
		model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
		return

	default:
		recvResults["status"] = constant.Failed
		writeResults["status"] = constant.Failed
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
	}
	return

}

func write(conn *websocket.Conn, wg *sync.WaitGroup, url string, body string, headers map[string][]string, wsConfig model.WsConfig, timeAfter <-chan time.Time, ticker *time.Ticker, writeResults map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()

	for {
		err := reconnection(conn, url, headers, wsConfig.RetryNum)
		if conn == nil {
			writeResults["err"] = err
			writeResults["is_stop"] = true
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			return
		}
		select {
		case <-timeAfter:
			writeResults["is_stop"] = true
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			return
		case <-ticker.C:
			bodyBytes := []byte(body)
			err = conn.WriteMessage(websocket.TextMessage, bodyBytes)
			writeResults["request_body"] = body
			writeResults["is_stop"] = false
			if err != nil {
				writeResults["err"] = err.Error()
				model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
				continue
			}
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
		}
	}

}

func read(conn *websocket.Conn, wg *sync.WaitGroup, url string, headers map[string][]string, wsConfig model.WsConfig, timeAfter <-chan time.Time, recvResults map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	for {
		log.Logger.Debug("11111111111111111111  .......................", conn)
		err := reconnection(conn, url, headers, wsConfig.RetryNum)
		if err != nil {
			log.Logger.Debug("errrrrrr:    ", err.Error())
			recvResults["err"] = err.Error()
			recvResults["is_stop"] = true
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		log.Logger.Debug("  .......................", conn)
		select {
		case <-timeAfter:
			recvResults["err"] = err
			recvResults["is_stop"] = true
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		default:
			log.Logger.Debug("conn:     ", conn)
			m, p, err := conn.ReadMessage()
			if err != nil {
				recvResults["err"] = err.Error()
				recvResults["is_stop"] = false
				model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
				break
			}
			recvResults["err"] = err
			recvResults["response_message_type"] = m
			recvResults["response_body"] = string(p)
			recvResults["is_stop"] = false
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
		}
	}
}

func reconnection(conn *websocket.Conn, url string, headers map[string][]string, retryNum int) (err error) {
	for i := 0; i < retryNum; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, headers)
		if conn != nil {
			log.Logger.Debug("con22222222222222222222:     ", conn)
			return
		}
	}
	return
}
