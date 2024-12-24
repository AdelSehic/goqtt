package server

import (
	"encoding/json"
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
)

type ConnReadJob struct {
	Conn       *Connection
	Buffer     []byte
	Recieved   int
	RemoteAddr string
}

func (job *ConnReadJob) Run() {
	message := &Message{}
	rawJson := job.Buffer[:job.Recieved]
	if err := json.Unmarshal(rawJson, message); err != nil {
		logger.Default.Error().Err(err).Msg("Error while unmarshalling JSON: " + string(rawJson))
		workers.GlobalPool.QueueJob(NewWriteJob(job.Conn, []byte("Invalid JSON recieved"), 0))
		return
	}
	message.Sender = job.Conn.ID

	response, _ := json.Marshal(message)
	switch message.Type {
	case EV_PING:
		response = []byte("Pong")
	case EV_ACKNOWLEDGE:
		job.Conn.AckChan <- struct{}{}
		response = []byte("Acknowledge recieved")
	case EV_SUBSCRIBE:
		if message.Topic == "" {
			break
		}
		workers.GlobalPool.QueueJob(&SubscribeJob{
			EventString: message.Topic,
			Conn:        job.Conn,
		})
		response = []byte("Subscribed to event")
	case EV_PUBLISH:
		workers.GlobalPool.QueueJob(&PublishJob{
			Conn: job.Conn,
			Msg:  message,
			Data: response,
		})
		response = []byte("Event queued for publishing")
	case EV_KEEPALIVE:
		logger.Default.Info().Msg(string(rawJson))
		return
	default:
		response = []byte("Invalid string")
	}
	workers.GlobalPool.QueueJob(NewWriteJob(job.Conn, response, 0))
	logger.Default.Info().Msg(string(rawJson))
	logger.HTTP.Info().Msg(string(rawJson))
}

func (job *ConnReadJob) Summary() string {
	return fmt.Sprintf("Recieving a message from %s ...", job.RemoteAddr)
}

func NewReadJob(conn *Connection) *ConnReadJob {
	logger.Default.Info().Msgf("New message from %s, starting read job ...", conn.Conn.RemoteAddr().String())
	job := &ConnReadJob{
		Conn:       conn,
		RemoteAddr: conn.Conn.RemoteAddr().String(),
	}
	job.Buffer = make([]byte, 1024)
	copy(job.Buffer, conn.buffer)
	job.Recieved = conn.recv
	return job
}
