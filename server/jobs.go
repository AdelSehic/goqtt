package server

import (
	"goqtt/logger"
	"net"
	"strings"
	"sync"
)

type ConnAcceptJob struct {
	Srv  *Server
	Conn *net.TCPConn
}

func (job *ConnAcceptJob) Run() {
	conn := &Connection{
		Conn:   job.Conn,
		ctx:    job.Srv.ctx,
		hertz:  60,
		buffer: make([]byte, 1024),
		lock:   &sync.Mutex{},
	}
	job.Srv.Conns[conn.Conn.RemoteAddr().String()] = conn
	go conn.HandleConnection()
	logger.Console.Info().Msgf("New connection established with %s", job.Conn.RemoteAddr().String())
}

type ConnReadJob struct {
	Buffer   []byte
	Recieved int
}

func (conn *ConnReadJob) Run() {
	message := string(conn.Buffer[:conn.Recieved])
	message = strings.Trim(message, "\r\n")
	logger.Console.Println(message)
}

func NewReadJob(conn *Connection) *ConnReadJob {
	logger.Console.Info().Msgf("New message from %s, starting read job ...", conn.Conn.RemoteAddr().String())
	job := &ConnReadJob{}
	job.Buffer = make([]byte, 1024)
	copy(job.Buffer, conn.buffer)
	job.Recieved = conn.recv
	return job
}

type ConnWriteJob struct {
	Conn       *net.TCPConn
	Buffer     []byte
}

func NewWriteJob(conn *net.TCPConn, data []byte) *ConnWriteJob {
	job := &ConnWriteJob{
		Conn: conn,
		Buffer: make([]byte, 1024),
	}
	copy(job.Buffer, data)
	return job
}

func (job *ConnWriteJob) Run() {
	if _, err := job.Conn.Write(job.Buffer); err != nil {
		logger.Console.Err(err).Msg("Error writing to connection!")
	}
}
