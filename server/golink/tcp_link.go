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
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	if tcp.TcpConfig == nil {
		return
	}
	tcp.TcpConfig.Init()

	timeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime) * time.Second)
	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Second)
	buf := make([]byte, 1024)
	results := make(map[string]interface{})
	results["uuid"] = tcp.Uuid.String()
	results["name"] = tcp.Name
	results["team_id"] = tcp.TeamId
	results["target_id"] = tcp.TargetId

	switch tcp.TcpConfig.ConnectType {
	case 1:

		wg := new(sync.WaitGroup)
		wg.Add(1)
		Read(wg, timeAfter, buf, conn, tcp, results, mongoCollection)
		wg.Add(1)
		Write(wg, timeAfter, ticker, conn, tcp, results, mongoCollection)
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

func Write(wg *sync.WaitGroup, timeAfter <-chan time.Time, ticker *time.Ticker, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()

	log.Logger.Debug("conn:    ", conn)
	for {
		select {
		case <-timeAfter:
			results["is_stop"] = true
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			return
		case <-ticker.C:
			msg := []byte(tcp.SendMessage)
			for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
				conn = client.NewTcpClient(tcp.Url)
				if conn != nil {
					break
				}
			}
			if conn == nil {
				for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
					conn = client.NewTcpClient(tcp.Url)
					if conn != nil {
						break
					}
				}
				if conn == nil {
					results["is_stop"] = true
					model.Insert(mongoCollection, results, middlewares.LocalIp)
					return
				}
			}
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

func Read(wg *sync.WaitGroup, timeAfter <-chan time.Time, buf []byte, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	log.Logger.Debug("conn:    ", len(buf), cap(buf))
	for {
		select {
		case <-timeAfter:
			results["is_stop"] = true
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			return
		default:
			if conn == nil {
				for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
					conn = client.NewTcpClient(tcp.Url)
					if conn != nil {
						break
					}
				}
				if conn == nil {
					results["is_stop"] = true
					model.Insert(mongoCollection, results, middlewares.LocalIp)
					return
				}
			}
			log.Logger.Debug("connnnnnnn:     ", conn)
			n, err := conn.Read(buf)
			log.Logger.Debug("开始读。。。。。。。。。。。。。。。。。11111111111111111111", err)
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
