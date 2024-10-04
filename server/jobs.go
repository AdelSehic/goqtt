package server

import (
	"fmt"
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
	logmsg := fmt.Sprintf("Connection established with %s", job.Conn.RemoteAddr())
	logger.Console.Info().Msg(logmsg)
	logger.HTTP.Info().Msg(logmsg)
}

func (job *ConnAcceptJob) Summary() string {
	return fmt.Sprintf("Recieving connection from %s", job.Conn.RemoteAddr())
}

type ConnReadJob struct {
	Buffer   []byte
	Recieved int
	RemoteAddr string
}

func (conn *ConnReadJob) Run() {
	message := string(conn.Buffer[:conn.Recieved])
	message = strings.Trim(message, "\r\n")
	logger.Console.Println(message)
}

func (job *ConnReadJob) Summary() string {
	return fmt.Sprintf("Recieving a message from %s ...", job.RemoteAddr)
}

func NewReadJob(conn *Connection) *ConnReadJob {
	logger.Console.Info().Msgf("New message from %s, starting read job ...", conn.Conn.RemoteAddr().String())
	job := &ConnReadJob{
		RemoteAddr: conn.Conn.RemoteAddr().String(),
	}
	job.Buffer = make([]byte, 1024)
	copy(job.Buffer, conn.buffer)
	job.Recieved = conn.recv
	return job
}

type ConnWriteJob struct {
	Conn   *net.TCPConn
	Buffer []byte
}

func NewWriteJob(conn *net.TCPConn, data []byte) *ConnWriteJob {
	job := &ConnWriteJob{
		Conn:   conn,
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

func (job *ConnWriteJob) Summary() string {
	return fmt.Sprintf("Writing message to %s ...", job.Conn.RemoteAddr())
}
