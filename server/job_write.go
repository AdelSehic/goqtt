package server

import (
	"fmt"
	"goqtt/logger"
	"net"
	"time"
)

type ConnWriteJob struct {
	Conn   *net.TCPConn
	Buffer []byte
	QoS    int8
}

func NewWriteJob(conn *net.TCPConn, data []byte, qos int8) *ConnWriteJob {
	job := &ConnWriteJob{
		Conn:   conn,
		Buffer: make([]byte, 1024),
		QoS:    qos,
	}
	copy(job.Buffer, data)
	return job
}

func (job *ConnWriteJob) Run() {
	logger.Console.Info().Msg(job.Summary())
	switch job.QoS {
	case 0:
		job.qos0write()
	case 1:
		job.qos1write()
	case 2:
		job.qos2write()
	}
}

func (job *ConnWriteJob) qos0write() {
	job.Conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := job.Conn.Write(job.Buffer); err != nil {
		logger.Console.Err(err).Msg("Error writing to connection!")
	}
}

func (job *ConnWriteJob) qos1write() {

}

func (job *ConnWriteJob) qos2write() {

}

func (job *ConnWriteJob) Summary() string {
	return fmt.Sprintf("Writing message %s to %s (QoS: %d) ...", string(job.Buffer), job.Conn.RemoteAddr(), job.QoS)
}
