package model

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPDetail struct {
	Timeout        int64           `json:"timeout"` // 请求超时时间
	Debug          string          `json:"debug"`   // 是否开启Debug模式
	Url            string          `json:"url"`
	SendMessage    string          `json:"send_message"`
	MessageType    string          `json:"message_type"` // "Binary"、"Text"、"Json"、"Xml"
	TcpConfig      *TcpConfig      `json:"tcp_config"`
	Configuration  *Configuration  `json:"configuration"`
	SqlVariable    *GlobalVariable `json:"sql_variable"`    // 全局变量
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
}

type TcpConfig struct {
	ConnectType         int32 `json:"connect_type"`           // 连接类型：1-长连接，2-短连接
	IsAutoSend          int32 `json:"is_auto_send"`           // 是否自动发送消息：0-非自动，1-自动
	ConnectTimeoutTime  int   `json:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int   `json:"retry_num"`              // 重连次数
	RetryInterval       int   `json:"retry_interval"`         // 重连间隔时间，单位：毫秒
	ConnectDurationTime int   `json:"connect_duration_time"`  // 连接持续时长
	SendMsgDurationTime int   `json:"send_msg_duration_time"` // 发送消息间隔时长
}

func (tcp TCPDetail) Send(debug string, debugMsg map[string]interface{}, mongoCollection *mongo.Collection) {
	var conn net.Conn
	var err error
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

	tcp.Url = strings.TrimSpace(tcp.Url)
	for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
		conn, err = NewTcpClient(tcp.Url)
		if conn != nil {
			connectionResults["status"] = true
			connectionResults["is_stop"] = false
			break
		}
	}

	if err != nil {
		recvResults["status"] = false
		writeResults["status"] = false
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		recvResults["response_body"] = err.Error()
		writeResults["request_body"] = err.Error()
		return
	}

	if tcp.TcpConfig == nil {
		recvResults["status"] = false
		writeResults["status"] = false
		recvResults["is_stop"] = true
		writeResults["is_stop"] = true
		recvResults["response_body"] = "tcpConfig is nil"
		writeResults["request_body"] = "tcpConfig is nil"
		return
	}
	tcp.TcpConfig.Init()

	readTimeAfter, writeTimeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second)
	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Millisecond)
	buf := make([]byte, 1024)

	switch tcp.TcpConfig.ConnectType {
	// 长连接
	case constant.LongConnection:
		adjustKey := fmt.Sprintf("TcpStatusChange:%s", debugMsg["uuid"])
		pubSub := SubscribeMsg(adjustKey)
		statusCh := pubSub.Channel()
		connChan := make(chan net.Conn, 2)
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go Read(wg, readTimeAfter, connChan, buf, conn, tcp, recvResults, mongoCollection, statusCh)
		wg.Add(1)
		go Write(wg, writeTimeAfter, connChan, ticker, conn, tcp, writeResults, connectionResults, mongoCollection, statusCh)
		wg.Wait()
	// 短连接
	case constant.ShortConnection:
		msg := []byte(tcp.SendMessage)
		_, err := conn.Write(msg)
		if err != nil {
			writeResults["status"] = false
			writeResults["request_body"] = err.Error()
		} else {
			writeResults["status"] = true
			writeResults["request_body"] = msg
		}
		writeResults["is_stop"] = true
		Insert(mongoCollection, writeResults, middlewares.LocalIp)
		n, err := conn.Read(buf)
		if err != nil {
			recvResults["status"] = false
			recvResults["response_body"] = err.Error()
		} else {
			recvResults["status"] = true
			recvResults["status"] = true
		}
		var result string
		if n != 0 {
			result = string(buf[:n])
		}
		recvResults["is_stop"] = true
		recvResults["response_body"] = result
		Insert(mongoCollection, recvResults, middlewares.LocalIp)
		return
	}
}

func NewTcpClient(url string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", url)
	return
}

func ReConnection(tcp TCPDetail, connChan chan net.Conn) (conn net.Conn, err error) {
	for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
		conn, err = NewTcpClient(tcp.Url)
		if conn != nil {
			for j := 0; j < 2; j++ {
				connChan <- conn
			}
			return
		}
	}
	return
}

func Write(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, ticker *time.Ticker, conn net.Conn, tcp TCPDetail, results, connectionResults map[string]interface{}, mongoCollection *mongo.Collection, statusCh <-chan *redis.Message) {
	defer wg.Done()
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if ticker != nil {
			ticker.Stop()
		}
	}()
	var err error

	tcpStatusChange := new(ConnectionStatusChange)
	switch tcp.TcpConfig.IsAutoSend {
	case constant.AutoConnectionSend:
		for {
			select {
			case <-timeAfter:
				results["status"] = true
				results["is_stop"] = true
				Insert(mongoCollection, results, middlewares.LocalIp)
				return
			case c := <-statusCh:

				_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
				if tcpStatusChange.Type == 1 {
					results["status"] = true
					results["is_stop"] = true
					Insert(mongoCollection, results, middlewares.LocalIp)
					return
				}
			case <-ticker.C:
				msg := []byte(tcp.SendMessage)
				if conn == nil {
					conn, err = ReConnection(tcp, connChan)
					results["status"] = false
					results["is_stop"] = true
					results["response_body"] = err.Error()
					Insert(mongoCollection, results, middlewares.LocalIp)
					break
				}
				select {
				case conn = <-connChan:
					_, err := conn.Write(msg)
					if err != nil {
						results["status"] = false
						results["is_stop"] = true
						results["response_body"] = err.Error()
						Insert(mongoCollection, results, middlewares.LocalIp)
						break
					}
					results["request_body"] = tcp.SendMessage
					if err != nil {
						results["status"] = false
						results["request_body"] = err.Error()
					} else {
						results["status"] = true
					}
					results["is_stop"] = false
					Insert(mongoCollection, results, middlewares.LocalIp)
					//log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
				default:
					_, err := conn.Write(msg)
					if err != nil {
						results["status"] = false
						results["is_stop"] = true
						results["response_body"] = err.Error()
						Insert(mongoCollection, results, middlewares.LocalIp)
						break
					}
					results["request_body"] = tcp.SendMessage
					if err != nil {
						results["status"] = false
						results["request_body"] = err.Error()
					} else {
						results["status"] = true
					}
					results["is_stop"] = false
					Insert(mongoCollection, results, middlewares.LocalIp)
					//log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
				}

			}
		}
	case constant.ConnectionAndSend:
		connectionResults["status"] = true
		connectionResults["is_stop"] = false
		Insert(mongoCollection, connectionResults, middlewares.LocalIp)
		for {
			select {
			case <-timeAfter:
				results["status"] = true
				results["is_stop"] = true
				Insert(mongoCollection, results, middlewares.LocalIp)
				return
			case c := <-statusCh:
				_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
				switch tcpStatusChange.Type {
				case constant.UnConnection:
					results["status"] = true
					results["is_stop"] = true
					Insert(mongoCollection, results, middlewares.LocalIp)
					return
				case constant.SendMessage:
					tcp.SendMessage = tcpStatusChange.Message
					msg := []byte(tcp.SendMessage)
					if conn == nil {
						conn, err = ReConnection(tcp, connChan)
						results["status"] = false
						results["is_stop"] = true
						results["response_body"] = err.Error()
						Insert(mongoCollection, results, middlewares.LocalIp)
						break
					}
					select {
					case conn = <-connChan:
						_, err := conn.Write(msg)
						if err != nil {
							results["status"] = false
							results["is_stop"] = true
							results["response_body"] = err.Error()
							Insert(mongoCollection, results, middlewares.LocalIp)
							break
						}
						results["request_body"] = tcp.SendMessage
						if err != nil {
							results["status"] = false
							results["request_body"] = err.Error()
						} else {
							results["status"] = true
						}
						results["is_stop"] = false
						Insert(mongoCollection, results, middlewares.LocalIp)
					default:
						_, err := conn.Write(msg)
						if err != nil {
							results["status"] = false
							results["is_stop"] = true
							results["response_body"] = err.Error()
							Insert(mongoCollection, results, middlewares.LocalIp)
							break
						}
						results["request_body"] = tcp.SendMessage
						if err != nil {
							results["status"] = false
							results["request_body"] = err.Error()
						} else {
							results["status"] = true
						}
						results["is_stop"] = false
						Insert(mongoCollection, results, middlewares.LocalIp)
					}

				}

			}
		}
	default:

		results["status"] = false
		results["is_stop"] = true
		results["response_body"] = err.Error()
		Insert(mongoCollection, results, middlewares.LocalIp)
		return
	}

}

func Read(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, buf []byte, conn net.Conn, tcp TCPDetail, results map[string]interface{}, mongoCollection *mongo.Collection, statusCh <-chan *redis.Message) {
	defer wg.Done()
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	var err error
	var tcpStatusChange = new(ConnectionStatusChange)
	for {
		select {
		case <-timeAfter:
			results["status"] = true
			results["is_stop"] = true
			Insert(mongoCollection, results, middlewares.LocalIp)
			return
		case c := <-statusCh:
			_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
			switch tcpStatusChange.Type {
			case 1:
				results["status"] = true
				results["is_stop"] = true
				Insert(mongoCollection, results, middlewares.LocalIp)
				return
			}
		default:
			if conn == nil {
				conn, err = ReConnection(tcp, connChan)
				if err != nil {
					results["status"] = false
					results["is_stop"] = true
					results["response_body"] = err.Error()
					Insert(mongoCollection, results, middlewares.LocalIp)
				}
			}
			select {
			case conn = <-connChan:
				n, err := conn.Read(buf)
				if err != nil {
					results["status"] = false
					results["response_body"] = err.Error()
					results["is_stop"] = false
					Insert(mongoCollection, results, middlewares.LocalIp)
					break
				} else {
					results["status"] = true
				}
				var msg string
				if n != 0 {
					msg = string(buf[:n])
				}
				results["response_body"] = msg
				results["is_stop"] = false
				Insert(mongoCollection, results, middlewares.LocalIp)
			default:
				n, err := conn.Read(buf)
				if err != nil {
					results["status"] = false
					results["response_body"] = err.Error()
					results["is_stop"] = false
					Insert(mongoCollection, results, middlewares.LocalIp)
					break
				} else {
					results["status"] = true
				}
				var msg string
				if n != 0 {
					msg = string(buf[:n])
				}
				results["response_body"] = msg
				results["is_stop"] = false
				Insert(mongoCollection, results, middlewares.LocalIp)
			}

		}
	}
}

func (tcpConfig *TcpConfig) Init() {
	if tcpConfig.RetryInterval == 0 {
		tcpConfig.RetryInterval = 1
	}
	if tcpConfig.ConnectTimeoutTime == 0 {
		tcpConfig.ConnectTimeoutTime = 1
	}
	if tcpConfig.RetryNum == 0 {
		tcpConfig.RetryNum = 1
	}
	if tcpConfig.ConnectDurationTime == 0 {
		tcpConfig.ConnectDurationTime = 1
	}
	if tcpConfig.SendMsgDurationTime == 0 {
		tcpConfig.SendMsgDurationTime = 1
	}
	if tcpConfig.RetryInterval == 0 {
		tcpConfig.RetryInterval = 1
	}
}
