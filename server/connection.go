package server

import (
	"context"
	"goqtt/logger"
	"goqtt/workers"
	"io"
	"net"
	"sync"
	"time"
)

type Connection struct {
	Conn   *net.TCPConn
	ctx    context.Context
	hertz  int
	buffer []byte
	recv   int
	lock   *sync.Mutex
}

func (conn *Connection) Lock() {
	conn.lock.Lock()
}

func (conn *Connection) Unlock() {
	conn.lock.Unlock()
}

func (conn *Connection) HandleConnection() {
	timeout := conn.hertz
	var err error
	conn.buffer = make([]byte, 1024)

	logger.Console.Info().Msgf("Opened connection to %s", conn.Conn.RemoteAddr().String())
	for {
		select {
		case <-conn.ctx.Done():
			conn.Conn.Close()
			logger.Console.Info().Msgf("Closed connection to %s (SIG)", conn.Conn.RemoteAddr().String())
			return
		default:
			conn.Conn.SetDeadline(time.Now().Add(time.Second))
			if conn.recv, err = conn.Conn.Read(conn.buffer); err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					if timeout <= 0 {
						logger.Console.Info().Msgf("Closed connection to %s (timeout)", conn.Conn.RemoteAddr().String())
						_ = conn.Conn.Close()
						return
					}
					timeout--
					continue
				}

				if err != io.EOF { // Handle non-EOF errors
					logger.Console.Err(err).Msgf("Connection %s dropped", conn.Conn.RemoteAddr().String())
				}

				_ = conn.Conn.Close()
				return
			}
			timeout = conn.hertz
			conn.Lock()
			workers.GlobalPool.QueueJob(NewReadJob(conn))
			conn.Unlock()
		}
	}
}
