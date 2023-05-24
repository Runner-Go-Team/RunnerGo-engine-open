package client

import (
	"fmt"
	"net"
)

func newTcpClient(ip string, port int) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil
	}
	return conn
}
