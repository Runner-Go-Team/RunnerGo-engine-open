package client

import (
	"net"
)

func NewTcpClient(url string) net.Conn {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		return nil
	}
	return conn
}
