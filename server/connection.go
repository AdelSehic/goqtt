package server

import "net"

type Connection interface {
	Send([]byte) error
	Read() ([]byte, error)
	Close()
}

type TCPConn struct {
	conn *net.TCPConn
}
