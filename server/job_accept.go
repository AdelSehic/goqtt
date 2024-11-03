package server

import (
	"fmt"
	"goqtt/logger"
	"net"
	"sync"
)

type ConnAcceptJob struct {
	Srv  *Server
	Conn *net.TCPConn
}

func (job *ConnAcceptJob) Run() {
	conn := &Connection{
		Conn:    job.Conn,
		ctx:     job.Srv.ctx,
		hertz:   60,
		buffer:  make([]byte, 1024),
		lock:    &sync.Mutex{},
		stop:    make(chan struct{}, 16),
		WG:      &sync.WaitGroup{},
		AckChan: make(chan struct{}, 16),
		Notify:  make([]string, 0),
	}
	go conn.HandleConnection()
	logmsg := fmt.Sprintf("Connection established with %s", job.Conn.RemoteAddr())
	logger.Console.Info().Msg(logmsg)
	logger.HTTP.Info().Msg(logmsg)
}

func (job *ConnAcceptJob) Summary() string {
	return fmt.Sprintf("Recieving connection from %s", job.Conn.RemoteAddr())
}
