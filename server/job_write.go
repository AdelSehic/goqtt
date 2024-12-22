package server

import (
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
	"time"
)

type ConnWriteJob struct {
	Conn   *Connection
	Buffer []byte
	QoS    int8
}

func NewWriteJob(conn *Connection, data []byte, qos int8) *ConnWriteJob {
	job := &ConnWriteJob{
		Conn:   conn,
		Buffer: make([]byte, 1024),
		QoS:    qos,
	}
	copy(job.Buffer, data)
	return job
}

func (job *ConnWriteJob) Run() {
	if !job.Conn.Active {
		return
	}
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
	job.Conn.Conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := job.Conn.Conn.Write(job.Buffer); err != nil {
		logger.Console.Err(err).Msg("Error writing to connection (QoS 0)")
		logger.Console.Debug().Stack().Err(err)
		job.Conn.stop <- struct{}{}
		job.Conn.Active = false
	}
}

func (job *ConnWriteJob) qos1write() {
	job.Conn.Conn.SetDeadline(time.Now().Add(1 * time.Second))
	if _, err := job.Conn.Conn.Write(job.Buffer); err != nil {
		logger.Console.Err(err).Msg("Conn unavalible, added to notifications")
		job.Conn.Notify = append(job.Conn.Notify, string(job.Buffer))
		job.Conn.stop <- struct{}{}
		return
	}

	select {
	case <-job.Conn.AckChan:
		logger.Console.Info().Msgf("Acknowledge recieved from %s", job.Conn.Conn.RemoteAddr())
		return
	case <-time.After(2 * time.Second):
		logger.Console.Info().Msgf("Have not recieved ACK from %s, resending ...", job.Conn.Conn.RemoteAddr())
	}
	workers.GlobalPool.QueueJob(job)
}

func (job *ConnWriteJob) qos2write() {

}

func (job *ConnWriteJob) Summary() string {
	return fmt.Sprintf("Writing message [%s] to [%s] (QoS: %d) ...", string(job.Buffer), job.Conn.Conn.RemoteAddr(), job.QoS)
}
