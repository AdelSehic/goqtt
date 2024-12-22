package server

import (
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
	"strconv"
	"strings"
)

type ConnReadJob struct {
	Conn       *Connection
	Buffer     []byte
	Recieved   int
	RemoteAddr string
}

func (job *ConnReadJob) Run() {
	message := string(job.Buffer[:job.Recieved])
	message = strings.Trim(message, "\r\n")
	fields := strings.Split(message, ",")
	var response []byte
	switch fields[0] {
	case EV_ACKNOWLEDGE:
		if len(fields) != 1 {
			break
		}
		job.Conn.AckChan <- struct{}{}
		response = []byte("Acknowledge recieved")
	case EV_SUBSCRIBE:
		if len(fields) != 2 {
			break
		}
		workers.GlobalPool.QueueJob(&SubscribeJob{
			EventString: fields[1],
			Conn:        job.Conn,
		})
		response = []byte("Subscribed to event!")
	case EV_PUBLISH:
		if len(fields) != 4 {
			break
		}
		qos, _ := strconv.Atoi(fields[3])
		workers.GlobalPool.QueueJob(&PublishJob{
			EventString: fields[1],
			Data:        fields[2],
			Conn:        job.Conn,
			QoS:         int8(qos),
		})
		response = []byte(fmt.Sprintf("Event queued for publishing (QoS: %d)", qos))
	default:
		response = []byte("Invalid string")
	}
	workers.GlobalPool.QueueJob(NewWriteJob(job.Conn, response, 0))
	logger.Console.Info().Msg(message)
	logger.HTTP.Info().Msg(message)
}

func (job *ConnReadJob) Summary() string {
	return fmt.Sprintf("Recieving a message from %s ...", job.RemoteAddr)
}

func NewReadJob(conn *Connection) *ConnReadJob {
	logger.Console.Info().Msgf("New message from %s, starting read job ...", conn.Conn.RemoteAddr().String())
	job := &ConnReadJob{
		Conn:       conn,
		RemoteAddr: conn.Conn.RemoteAddr().String(),
	}
	job.Buffer = make([]byte, 1024)
	copy(job.Buffer, conn.buffer)
	job.Recieved = conn.recv
	return job
}
