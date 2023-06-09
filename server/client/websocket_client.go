package client

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

func WebSocketRequest(recvResults, writeResults map[string]interface{}, mongoCollection *mongo.Collection, url string, body string, headers map[string][]string, wsConfig model.WsConfig) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	var conn *websocket.Conn

	writeResults["type"] = "send"
	recvResults["type"] = "recv"
	for i := 0; i < wsConfig.RetryNum; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, headers)
		if conn != nil {
			break
		}
	}
	if err != nil || conn == nil {
		recvResults["err"] = err
		writeResults["err"] = err
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
	switch wsConfig.ConnectType {
	case 1:
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func(wsWg *sync.WaitGroup) {
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
						writeResults["err"] = err
						writeResults["is_stop"] = true
						model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
						return
					}

				}
				select {
				case <-writeTimeAfter:
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

		}(wg)
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
						recvResults["err"] = err.Error()
						recvResults["is_stop"] = true
						model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
						return
					}

				}
				select {
				case <-readTimeAfter:
					recvResults["err"] = err
					recvResults["is_stop"] = true
					model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
					return
				default:
					m, p, connectErr := conn.ReadMessage()
					if connectErr != nil {
						recvResults["err"] = connectErr.Error()
						recvResults["is_stop"] = false
						model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
						break
					}
					recvResults["err"] = connectErr
					recvResults["response_message_type"] = m
					recvResults["response_body"] = string(p)
					recvResults["is_stop"] = false
					model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
				}
			}
		}()
		wg.Wait()
	case 2:
		if conn == nil {
			recvResults["err"] = err
			writeResults["err"] = err
			ticker.Stop()
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
			writeResults["err"] = err.Error()
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
			model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
			model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
			return
		}
		recvResults["response_message_type"] = m
		recvResults["response_body"] = string(p)
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
		model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
		return

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
