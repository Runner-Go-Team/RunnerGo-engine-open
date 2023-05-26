package golink

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"net"
	"time"
)

func TcpConnection(tcp model.TCP) {

	var conn net.Conn
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	switch tcp.TcpConfig.ConnectType {
	case 1:
		timeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime) * time.Second)
		ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Second)
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
						n, err := conn.Write([]byte(tcp.SendMessage))
						if err != nil {
							log.Logger.Debug("tcp 写入消息失败:", err.Error())
							continue
						}
						log.Logger.Debug("tcp写入消息：%d, %s", n, tcp.SendMessage)
					default:
						buf := []byte{}
						n, err := conn.Read(buf)
						if err != nil {
							log.Logger.Debug("tcp 读取消息失败:", err.Error())
							continue
						}
						log.Logger.Debug("读取消息：%d, %s", n, string(buf))
					}
				}

			}
		}
	case 2:

	}
}
