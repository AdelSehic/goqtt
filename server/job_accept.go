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
		Conn:  job.Conn,
		ctx:   job.Srv.ctx,
		hertz: 60,
	}
	go conn.HandleConnection()
	logmsg := fmt.Sprintf("Connection established with %s", job.Conn.RemoteAddr())
	logger.Default.Info().Msg(logmsg)
	logger.HTTP.Info().Msg(logmsg)
}

func (conn *Connection) initFreshConn() {
	conn.buffer = make([]byte, 1024)
	conn.lock = &sync.Mutex{}
	conn.stop = make(chan struct{}, 16)
	conn.WG = &sync.WaitGroup{}
	conn.AckChan = make(chan struct{}, 16)
	conn.Notify = make([]string, 0)
}

func (job *ConnAcceptJob) Summary() string {
	return fmt.Sprintf("Recieving connection from %s", job.Conn.RemoteAddr())
}
