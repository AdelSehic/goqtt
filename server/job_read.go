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

func (conn *ConnReadJob) Run() {
	message := string(conn.Buffer[:conn.Recieved])
	message = strings.Trim(message, "\r\n")
	fields := strings.Split(message, ",")
	response := []byte("Invalid message string")
	switch fields[0] {
	case EV_SUBSCRIBE:
		if len(fields) == 2 {
			workers.GlobalPool.QueueJob(&SubscribeJob{
				EventString: fields[1],
				Conn:        conn.Conn,
			})
			response = []byte("Subscribed to event!")
		}
	case EV_PUBLISH:
		if len(fields) == 4 {
			qos, _ := strconv.Atoi(fields[3])
			workers.GlobalPool.QueueJob(&PublishJob{
				EventString: fields[1],
				Data:        fields[2],
				Conn:        conn.Conn,
				QoS:         int8(qos),
			})
			response = []byte("Event published!")
		}
	}
	workers.GlobalPool.QueueJob(NewWriteJob(conn.Conn.Conn, response, 0))
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
