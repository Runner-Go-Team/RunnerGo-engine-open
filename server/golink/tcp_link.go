package golink

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"time"
)

func TcpConnection(tcp model.TCP, mongoCollection *mongo.Collection) {

	var conn net.Conn
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	timeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime) * time.Second)
	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Millisecond)
	buf := make([]byte, 1024)
	results := make(map[string]interface{})
	results["uuid"] = tcp.Uuid.String()
	results["name"] = tcp.Name
	results["team_id"] = tcp.TeamId
	results["target_id"] = tcp.TargetId
	switch tcp.TcpConfig.ConnectType {
	case 1:
	Reconnect:
		for {
			select {
			case <-timeAfter:
				return
			default:
				conn = client.NewTcpClient(tcp.Url)
				if conn == nil {
					for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
						conn = client.NewTcpClient(tcp.Url)
						if conn != nil {
							break
						}
					}
				}
				if conn == nil {
					goto Reconnect
				}
				for {
					if conn == nil {
						goto Reconnect
					}
					select {
					case <-ticker.C:
						msg := []byte(tcp.SendMessage)
						_, err := conn.Write(msg)
						if err != nil {
							log.Logger.Debug("tcp 写入消息失败:", err.Error())
						}
						results["request_body"] = msg
						if err != nil {
							results["send_status"] = false
							results["send_err"] = err.Error()
						} else {
							results["send_status"] = true
							results["send_err"] = err
						}

						model.Insert(mongoCollection, results, middlewares.LocalIp)
						//fmt.Println(fmt.Sprintf("tcp写入消息：%d, %s", n, tcp.SendMessage))
						//log.Logger.Debug("tcp写入消息：%d, %s", n, tcp.SendMessage)
					default:
						n, err := conn.Read(buf[:])
						if err != nil {
							log.Logger.Debug("tcp 读取消息失败:", err.Error())
						}
						msg := string(buf[n])
						results["response_body"] = msg
						if err != nil {
							results["recv_status"] = false
							results["recv_err"] = err.Error()
						} else {
							results["recv_status"] = true
							results["recv_err"] = err
						}
						model.Insert(mongoCollection, results, middlewares.LocalIp)
						//fmt.Println(fmt.Sprintf("读取消息：%d, %s", n, string(buf)))
						log.Logger.Debug("读取消息：%d, %s", n, string(buf[n]))
					}
				}

			}
		}
	case 2:
		conn = client.NewTcpClient(tcp.Url)
		if conn == nil {
			for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
				conn = client.NewTcpClient(tcp.Url)
				if conn != nil {
					break
				}
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
