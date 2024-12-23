package server

import (
	"context"
	"goqtt/config"
	"goqtt/logger"
	"goqtt/workers"
	"net"
	"sync"
	"time"
)

type Server struct {
	Wg       *sync.WaitGroup
	ctx      context.Context
	Listener *net.TCPListener
	Stop     context.CancelFunc
}

func NewServer(cfg *config.Connector) *Server {

	addr, err := net.ResolveTCPAddr("tcp", cfg.Port)
	if err != nil {
		logger.Default.Err(err).Msg("Couldn't resolve port")
		return nil
	}

	lsn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Default.Err(err).Msg("Couldn't listen on provided address")
		return nil
	}

	ctx, stop := context.WithCancel(context.Background())
	return &Server{
		Listener: lsn,
		Wg:       &sync.WaitGroup{},
		ctx:      ctx,
		Stop:     stop,
	}
}

func (srv *Server) Start() {
	PoolInit()
	EventsInit()
	for {
		select {
		case <-srv.ctx.Done():
			return
		default:
			srv.Listener.SetDeadline(time.Now().Add(time.Second * 1))
			conn, err := srv.Listener.AcceptTCP()
			if err != nil {
				continue
			}
			workers.GlobalPool.QueueJob(&ConnAcceptJob{
				Srv: srv,
				Conn: conn,
			})
		}
	}
}
