package golink

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"sync"
	"time"
)

func TcpConnection(tcp model.TCP, mongoCollection *mongo.Collection) {
	var conn net.Conn
	conn = client.NewTcpClient(tcp.Url)
	if conn == nil {
		for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
			conn = client.NewTcpClient(tcp.Url)
			if conn != nil {
				break
			}
		}
		if conn == nil {
			return
		}
	}

	if tcp.TcpConfig == nil {
		return
	}
	tcp.TcpConfig.Init()

	readTimeAfter, writeTimeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second)
	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Millisecond)
	buf := make([]byte, 1024)
	results := make(map[string]interface{})
	results["uuid"] = tcp.Uuid.String()
	results["name"] = tcp.Name
	results["team_id"] = tcp.TeamId
	results["target_id"] = tcp.TargetId

	switch tcp.TcpConfig.ConnectType {
	case 1:
		connChan := make(chan net.Conn, 2)
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go Read(wg, readTimeAfter, connChan, buf, conn, tcp, results, mongoCollection)
		wg.Add(1)
		go Write(wg, writeTimeAfter, connChan, ticker, conn, tcp, results, mongoCollection)
		wg.Wait()
	case 2:
		conn = client.NewTcpClient(tcp.Url)
		if conn == nil {
			for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
				conn = client.NewTcpClient(tcp.Url)
				if conn != nil {
					break
				}
			}
			if conn == nil {
				return
			}
		}
		msg := []byte(tcp.SendMessage)
		_, err := conn.Write(msg)
		results["request_body"] = msg
		if err != nil {
			results["send_status"] = false
			results["send_err"] = err.Error()
		} else {
			results["send_status"] = true
			results["send_err"] = err
		}
		n, err := conn.Read(buf)
		if err != nil {
			results["recv_status"] = false
			results["recv_err"] = err.Error()
		} else {
			results["recv_status"] = true
			results["recv_err"] = err
		}
		var result string
		if n != 0 {
			result = string(buf[:n])
		}
		results["response_body"] = result
		model.Insert(mongoCollection, results, middlewares.LocalIp)
		log.Logger.Debug("接收到数据: %s", result)
		return
	}
}

func ReConnection(conn net.Conn, tcp model.TCP, connChan chan net.Conn) {
	for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
		conn = client.NewTcpClient(tcp.Url)
		if conn != nil {
			for j := 0; j < 2; j++ {
				connChan <- conn
			}

			break
		}
	}
}

func Write(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, ticker *time.Ticker, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	for {
		select {
		case <-timeAfter:
			results["is_stop"] = true
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			log.Logger.Debug("停止写")
			return
		case <-ticker.C:
			msg := []byte(tcp.SendMessage)
			if conn == nil {
				ReConnection(conn, tcp, connChan)
			}
			select {
			case conn = <-connChan:
				_, err := conn.Write(msg)
				if err != nil {
					log.Logger.Debug("tcp 写入消息失败:", err.Error())
				}
				results["request_body"] = tcp.SendMessage
				if err != nil {
					results["send_status"] = false
					results["send_err"] = err.Error()
				} else {
					results["send_status"] = true
					results["send_err"] = err
				}
				results["type"] = "send"
				results["is_stop"] = false
				model.Insert(mongoCollection, results, middlewares.LocalIp)
				log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
			default:
				_, err := conn.Write(msg)
				if err != nil {
					log.Logger.Debug("tcp 写入消息失败:", err.Error())
				}
				results["request_body"] = tcp.SendMessage
				if err != nil {
					results["send_status"] = false
					results["send_err"] = err.Error()
				} else {
					results["send_status"] = true
					results["send_err"] = err
				}
				results["type"] = "send"
				results["is_stop"] = false
				model.Insert(mongoCollection, results, middlewares.LocalIp)
				log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
			}

		}
	}

}

func Read(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, buf []byte, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	for {
		select {
		case <-timeAfter:
			results["is_stop"] = true
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			log.Logger.Debug("停止读")
			return
		default:
			if conn == nil {
				ReConnection(conn, tcp, connChan)
			}
			select {
			case conn = <-connChan:

				if conn == nil {
					results["is_stop"] = true
					model.Insert(mongoCollection, results, middlewares.LocalIp)
					return
				}
				n, err := conn.Read(buf)
				if err != nil {
					results["recv_status"] = false
					results["recv_err"] = err.Error()
				} else {
					results["recv_status"] = true
					results["recv_err"] = err
				}
				results["type"] = "recv"
				var msg string
				if n != 0 {
					msg = string(buf[:n])
				}
				results["response_body"] = msg
				results["is_stop"] = false
				model.Insert(mongoCollection, results, middlewares.LocalIp)
				log.Logger.Debug(fmt.Sprintf("tcp读消息: %s", msg))
			default:
				if conn == nil {
					results["is_stop"] = true
					model.Insert(mongoCollection, results, middlewares.LocalIp)
					return
				}
				n, err := conn.Read(buf)
				if err != nil {
					results["recv_status"] = false
					results["recv_err"] = err.Error()
				} else {
					results["recv_status"] = true
					results["recv_err"] = err
				}
				results["type"] = "recv"
				var msg string
				if n != 0 {
					msg = string(buf[:n])
				}
				results["response_body"] = msg
				results["is_stop"] = false
				model.Insert(mongoCollection, results, middlewares.LocalIp)
				log.Logger.Debug(fmt.Sprintf("tcp读消息: %s", msg))
			}

		}
	}
}
