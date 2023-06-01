package golink

import (
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
	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Millisecond)
	buf := []byte{}
	results := make(map[string]interface{})
	results["uuid"] = tcp.Uuid.String()
	results["name"] = tcp.Name
	results["team_id"] = tcp.TeamId
	results["target_id"] = tcp.TargetId

	switch tcp.TcpConfig.ConnectType {
	case 1:
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go Write(wg, timeAfter, ticker, conn, tcp, results, mongoCollection)
		wg.Add(1)
		go Read(wg, timeAfter, buf, conn, tcp, results, mongoCollection)
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
		for {
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
			n, err := conn.Read(buf[:])
			if err != nil {
				results["recv_status"] = false
				results["recv_err"] = err.Error()
			} else {
				results["recv_status"] = true
				results["recv_err"] = err
			}
			results["response_body"] = string(buf[n])
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			return
		}
	}
}

func Write(wg *sync.WaitGroup, timeAfter <-chan time.Time, ticker *time.Ticker, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	for {
		select {
		case <-timeAfter:
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
			model.Insert(mongoCollection, results, middlewares.LocalIp)
			log.Logger.Debug("tcp写入消息: %s", tcp.SendMessage)

		}
	}

}

func Read(wg *sync.WaitGroup, timeAfter <-chan time.Time, buf []byte, conn net.Conn, tcp model.TCP, results map[string]interface{}, mongoCollection *mongo.Collection) {
	defer wg.Done()
	for {
		log.Logger.Debug("开始读。。。。。。。。。。。。。。。。。")
		select {
		case <-timeAfter:
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
					return
				}
			}
			n, err := conn.Read(buf[:])
			if err != nil {
				results["recv_status"] = false
				results["recv_err"] = err.Error()
			} else {
				results["recv_status"] = true
				results["recv_err"] = err
			}
			results["type"] = "recv"
			if n == 0 {
				results["response_body"] = ""
			} else {
				results["response_body"] = string(buf[n])
			}

			model.Insert(mongoCollection, results, middlewares.LocalIp)
		}
	}
}
