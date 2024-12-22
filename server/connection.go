package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Connection struct {
	Conn    *net.TCPConn
	ctx     context.Context
	hertz   int
	buffer  []byte
	recv    int
	ID      string
	stop    chan struct{}
	Active  bool
	lock    *sync.Mutex
	WG      *sync.WaitGroup
	AckChan chan struct{}
	Notify  []string
}

func (conn *Connection) Lock() {
	conn.lock.Lock()
}

func (conn *Connection) Unlock() {
	conn.lock.Unlock()
}

func (conn *Connection) Close() {
	conn.stop <- struct{}{}
	conn.Conn.Close()
}

func (conn *Connection) HandleConnection() {
	defer conn.Close()
	timeout := conn.hertz
	var err error

	if err := conn.getID(); err != nil {
		logger.Console.Err(err).Msg("Refusing client connection")
		return
	}

	if ConnectionPool.ConnExists(conn.ID) {
		ConnectionPool.CloseConn(conn.ID)
		conn = ConnectionPool.Reconn(conn)
	} else {
		workers.GlobalPool.QueueJob(NewWriteJob(conn, []byte("Device registered!\n"), 0))
		ConnectionPool.NewConn(conn)
	}

	// clear the stop channel
	for len(conn.stop) > 0 {
		<-conn.stop
	}

	conn.Active = true
	logger.Console.Info().Msgf("Opened connection to %s (%s)", conn.Conn.RemoteAddr().String(), conn.ID)
	for {
		select {
		case <-conn.ctx.Done():
			logger.Console.Info().Msgf("Closed connection to %s (program shutdown)", conn.Conn.RemoteAddr().String())
			return
		case <-conn.stop:
			conn.Lock()
			conn.Active = false
			conn.Unlock()
			logger.Console.Info().Msgf("Closed connection to %s (connection interrupt)", conn.Conn.RemoteAddr().String())
			return
		default:
			conn.Conn.SetDeadline(time.Now().Add(time.Second))
			if conn.recv, err = conn.Conn.Read(conn.buffer); err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					if timeout <= 0 {
						logger.Console.Info().Msgf("Closed connection to %s (timeout)", conn.Conn.RemoteAddr().String())
						conn.Active = false
						return
					}
					timeout--
					continue
				}
				if err != io.EOF { // Handle non-EOF errors
					logger.Console.Err(err).Msgf("Connection %s dropped", conn.Conn.RemoteAddr().String())
				} else { // or handle EOF
					logger.Console.Info().Msgf("Connection %s closed by client", conn.Conn.RemoteAddr().String())
					logger.HTTP.Info().Msgf("Connection %s closed by client", conn.Conn.RemoteAddr().String())
				}
				return
			}
			timeout = conn.hertz
			conn.Lock()
			workers.GlobalPool.QueueJob(NewReadJob(conn))
			conn.Unlock()
			// workers.GlobalPool.QueueJob(NewWriteJob(conn.Conn, []byte("Thanks for visiting my server!\r\n")))
		}
	}
}

func (conn *Connection) getID() error {
	reader := bufio.NewReader(conn.Conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading client ID: %v", err)
	}

	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) != 2 || parts[0] != "ClientID" {
		return errors.New("invalid client ID header received")
	}

	conn.ID = parts[1]
	return nil
}
