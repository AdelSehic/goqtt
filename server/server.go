package server

import (
	"context"
	"fmt"
	"goqtt/config"
	"goqtt/logger"
	"net"
	"sync"
	"time"
)

type Server struct {
	Wg       *sync.WaitGroup
	ctx      context.Context
	Listener *net.TCPListener
	Stop     context.CancelFunc
	Conns    map[string]Connection
}

func NewServer(cfg *config.Connector) *Server {

	addr, err := net.ResolveTCPAddr("tcp", cfg.Port)
	if err != nil {
		logger.Console.Err(err).Msg("Couldn't resolve port")
		return nil
	}

	lsn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Console.Err(err).Msg("Couldn't listen on provided address")
		return nil
	}

	ctx, stop := context.WithCancel(context.Background())
	return &Server{
		Listener: lsn,
		Wg:       &sync.WaitGroup{},
		ctx:      ctx,
		Stop:     stop,
		Conns:    make(map[string]Connection),
	}
}

func (srv *Server) Start() {
	for {
		select {
		case <-srv.ctx.Done():
			return
		default:
			srv.Listener.SetDeadline(time.Now().Add(time.Second * 1))
			conn, err := srv.Listener.AcceptTCP()
			if err != nil {
				logger.Console.Info().Msg("No new connections")
				continue
			}
			fmt.Println(conn.RemoteAddr().String())
		}
	}
}
