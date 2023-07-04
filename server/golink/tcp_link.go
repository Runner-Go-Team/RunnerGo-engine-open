package golink

//func TcpConnection(tcp model.TCPDetail, mongoCollection *mongo.Collection) {
//	var conn net.Conn
//	var err error
//	connectionResults, recvResults, writeResults := make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{})
//
//	recvResults["type"] = "recv"
//	recvResults["uuid"] = tcp.Uuid.String()
//	recvResults["name"] = tcp.Name
//	recvResults["team_id"] = tcp.TeamId
//	recvResults["target_id"] = tcp.TargetId
//
//	writeResults["type"] = "send"
//	writeResults["uuid"] = tcp.Uuid.String()
//	writeResults["name"] = tcp.Name
//	writeResults["team_id"] = tcp.TeamId
//	writeResults["target_id"] = tcp.TargetId
//
//	connectionResults["type"] = "connection"
//	connectionResults["uuid"] = tcp.Uuid.String()
//	connectionResults["name"] = tcp.Name
//	connectionResults["team_id"] = tcp.TeamId
//	connectionResults["target_id"] = tcp.TargetId
//
//	for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
//		conn, err = client.NewTcpClient(tcp.Url)
//		if conn != nil {
//			connectionResults["status"] = true
//			connectionResults["is_stop"] = false
//			break
//		}
//	}
//
//	if err != nil {
//		recvResults["status"] = false
//		writeResults["status"] = false
//		recvResults["is_stop"] = true
//		writeResults["is_stop"] = true
//		recvResults["response_body"] = err.Error()
//		writeResults["request_body"] = err.Error()
//		return
//	}
//
//	if tcp.TcpConfig == nil {
//		recvResults["status"] = false
//		writeResults["status"] = false
//		recvResults["is_stop"] = true
//		writeResults["is_stop"] = true
//		recvResults["response_body"] = "tcpConfig is nil"
//		writeResults["request_body"] = "tcpConfig is nil"
//		return
//	}
//	tcp.TcpConfig.Init()
//
//	readTimeAfter, writeTimeAfter := time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second), time.After(time.Duration(tcp.TcpConfig.ConnectDurationTime)*time.Second)
//	ticker := time.NewTicker(time.Duration(tcp.TcpConfig.SendMsgDurationTime) * time.Millisecond)
//	buf := make([]byte, 1024)
//
//	switch tcp.TcpConfig.ConnectType {
//	// 长连接
//	case constant.LongConnection:
//		adjustKey := fmt.Sprintf("TcpStatusChange:%s", tcp.Uuid.String())
//		pubSub := model.SubscribeMsg(adjustKey)
//		statusCh := pubSub.Channel()
//		connChan := make(chan net.Conn, 2)
//		wg := new(sync.WaitGroup)
//		wg.Add(1)
//		go Read(wg, readTimeAfter, connChan, buf, conn, tcp, recvResults, mongoCollection, statusCh)
//		wg.Add(1)
//		go Write(wg, writeTimeAfter, connChan, ticker, conn, tcp, writeResults, connectionResults, mongoCollection, statusCh)
//		wg.Wait()
//	// 短连接
//	case constant.ShortConnection:
//		msg := []byte(tcp.SendMessage)
//		_, err := conn.Write(msg)
//		writeResults["request_body"] = msg
//		if err != nil {
//			writeResults["status"] = false
//			writeResults["request_body"] = err.Error()
//		} else {
//			writeResults["status"] = true
//			writeResults["send_err"] = err
//		}
//		writeResults["is_stop"] = true
//		model.Insert(mongoCollection, writeResults, middlewares.LocalIp)
//		n, err := conn.Read(buf)
//		if err != nil {
//			recvResults["status"] = false
//			recvResults["response_body"] = err.Error()
//		} else {
//			recvResults["status"] = true
//			recvResults["status"] = true
//		}
//		var result string
//		if n != 0 {
//			result = string(buf[:n])
//		}
//		recvResults["is_stop"] = true
//		recvResults["response_body"] = result
//		model.Insert(mongoCollection, recvResults, middlewares.LocalIp)
//		return
//	}
//}
//
//func ReConnection(tcp model.TCPDetail, connChan chan net.Conn) (conn net.Conn, err error) {
//	for i := 0; i < tcp.TcpConfig.RetryNum; i++ {
//		conn, err = client.NewTcpClient(tcp.Url)
//		if conn != nil {
//			for j := 0; j < 2; j++ {
//				connChan <- conn
//			}
//			return
//		}
//	}
//	return
//}
//
//func Write(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, ticker *time.Ticker, conn net.Conn, tcp model.TCPDetail, results, connectionResults map[string]interface{}, mongoCollection *mongo.Collection, statusCh <-chan *redis.Message) {
//	defer wg.Done()
//	defer func() {
//		if conn != nil {
//			conn.Close()
//		}
//		if ticker != nil {
//			ticker.Stop()
//		}
//	}()
//	var err error
//
//	tcpStatusChange := new(model.ConnectionStatusChange)
//	switch tcp.TcpConfig.IsAutoSend {
//	case constant.AutoConnectionSend:
//		for {
//			select {
//			case <-timeAfter:
//				results["status"] = true
//				results["is_stop"] = true
//				model.Insert(mongoCollection, results, middlewares.LocalIp)
//				return
//			case c := <-statusCh:
//
//				_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
//				if tcpStatusChange.Type == 1 {
//					results["status"] = true
//					results["is_stop"] = true
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					return
//				}
//			case <-ticker.C:
//				msg := []byte(tcp.SendMessage)
//				if conn == nil {
//					conn, err = ReConnection(tcp, connChan)
//					results["status"] = false
//					results["is_stop"] = true
//					results["response_body"] = err.Error()
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					break
//				}
//				select {
//				case conn = <-connChan:
//					_, err := conn.Write(msg)
//					if err != nil {
//						results["status"] = false
//						results["is_stop"] = true
//						results["response_body"] = err.Error()
//						model.Insert(mongoCollection, results, middlewares.LocalIp)
//						break
//					}
//					results["request_body"] = tcp.SendMessage
//					if err != nil {
//						results["status"] = false
//						results["request_body"] = err.Error()
//					} else {
//						results["status"] = true
//					}
//					results["is_stop"] = false
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
//				default:
//					_, err := conn.Write(msg)
//					if err != nil {
//						results["status"] = false
//						results["is_stop"] = true
//						results["response_body"] = err.Error()
//						model.Insert(mongoCollection, results, middlewares.LocalIp)
//						break
//					}
//					results["request_body"] = tcp.SendMessage
//					if err != nil {
//						results["status"] = false
//						results["request_body"] = err.Error()
//					} else {
//						results["status"] = true
//					}
//					results["is_stop"] = false
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					log.Logger.Debug(fmt.Sprintf("tcp写入消息: %s", tcp.SendMessage))
//				}
//
//			}
//		}
//	case constant.ConnectionAndSend:
//		connectionResults["status"] = true
//		connectionResults["is_stop"] = false
//		model.Insert(mongoCollection, connectionResults, middlewares.LocalIp)
//		for {
//			select {
//			case <-timeAfter:
//				results["status"] = true
//				results["is_stop"] = true
//				model.Insert(mongoCollection, results, middlewares.LocalIp)
//				return
//			case c := <-statusCh:
//				_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
//				switch tcpStatusChange.Type {
//				case constant.UnConnection:
//					results["status"] = true
//					results["is_stop"] = true
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					return
//				case constant.SendMessage:
//					tcp.SendMessage = tcpStatusChange.Message
//					msg := []byte(tcp.SendMessage)
//					if conn == nil {
//						conn, err = ReConnection(tcp, connChan)
//						results["status"] = false
//						results["is_stop"] = true
//						results["response_body"] = err.Error()
//						model.Insert(mongoCollection, results, middlewares.LocalIp)
//						break
//					}
//					select {
//					case conn = <-connChan:
//						_, err := conn.Write(msg)
//						if err != nil {
//							results["status"] = false
//							results["is_stop"] = true
//							results["response_body"] = err.Error()
//							model.Insert(mongoCollection, results, middlewares.LocalIp)
//							break
//						}
//						results["request_body"] = tcp.SendMessage
//						if err != nil {
//							results["status"] = false
//							results["request_body"] = err.Error()
//						} else {
//							results["status"] = true
//						}
//						results["is_stop"] = false
//						model.Insert(mongoCollection, results, middlewares.LocalIp)
//					default:
//						_, err := conn.Write(msg)
//						if err != nil {
//							results["status"] = false
//							results["is_stop"] = true
//							results["response_body"] = err.Error()
//							model.Insert(mongoCollection, results, middlewares.LocalIp)
//							break
//						}
//						results["request_body"] = tcp.SendMessage
//						if err != nil {
//							results["status"] = false
//							results["request_body"] = err.Error()
//						} else {
//							results["status"] = true
//						}
//						results["is_stop"] = false
//						model.Insert(mongoCollection, results, middlewares.LocalIp)
//					}
//
//				}
//
//			}
//		}
//	default:
//
//		results["status"] = false
//		results["is_stop"] = true
//		results["response_body"] = err.Error()
//		model.Insert(mongoCollection, results, middlewares.LocalIp)
//		return
//	}
//
//}
//
//func Read(wg *sync.WaitGroup, timeAfter <-chan time.Time, connChan chan net.Conn, buf []byte, conn net.Conn, tcp model.TCPDetail, results map[string]interface{}, mongoCollection *mongo.Collection, statusCh <-chan *redis.Message) {
//	defer wg.Done()
//	defer func() {
//		if conn != nil {
//			conn.Close()
//		}
//	}()
//	var err error
//	var tcpStatusChange = new(model.ConnectionStatusChange)
//	for {
//		select {
//		case <-timeAfter:
//			results["status"] = true
//			results["is_stop"] = true
//			model.Insert(mongoCollection, results, middlewares.LocalIp)
//			return
//		case c := <-statusCh:
//			_ = json.Unmarshal([]byte(c.Payload), tcpStatusChange)
//			switch tcpStatusChange.Type {
//			case 1:
//				results["status"] = true
//				results["is_stop"] = true
//				model.Insert(mongoCollection, results, middlewares.LocalIp)
//				return
//			}
//		default:
//			if conn == nil {
//				conn, err = ReConnection(tcp, connChan)
//				if err != nil {
//					results["status"] = false
//					results["is_stop"] = true
//					results["response_body"] = err.Error()
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//				}
//			}
//			select {
//			case conn = <-connChan:
//				n, err := conn.Read(buf)
//				if err != nil {
//					results["status"] = false
//					results["response_body"] = err.Error()
//					results["is_stop"] = false
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					break
//				} else {
//					results["status"] = true
//				}
//				var msg string
//				if n != 0 {
//					msg = string(buf[:n])
//				}
//				results["response_body"] = msg
//				results["is_stop"] = false
//				model.Insert(mongoCollection, results, middlewares.LocalIp)
//			default:
//				n, err := conn.Read(buf)
//				if err != nil {
//					results["status"] = false
//					results["response_body"] = err.Error()
//					results["is_stop"] = false
//					model.Insert(mongoCollection, results, middlewares.LocalIp)
//					break
//				} else {
//					results["status"] = true
//				}
//				var msg string
//				if n != 0 {
//					msg = string(buf[:n])
//				}
//				results["response_body"] = msg
//				results["is_stop"] = false
//				model.Insert(mongoCollection, results, middlewares.LocalIp)
//			}
//
//		}
//	}
//}
