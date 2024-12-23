package server

import (
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
	"sync"
)

var ConnectionPool *ConnPool

type ConnPool struct {
	Connections map[string]*Connection
	mtx         *sync.Mutex
}

func PoolInit() {
	ConnectionPool = &ConnPool{
		Connections: make(map[string]*Connection),
		mtx:         &sync.Mutex{},
	}
}

func (p *ConnPool) ConnExists(id string) bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, exists := p.Connections[id]
	return exists
}

func (p *ConnPool) NewConn(conn *Connection) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	conn.initFreshConn()
	p.Connections[conn.ID] = conn
}

func (p *ConnPool) Reconn(conn *Connection) *Connection {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.Connections[conn.ID].Conn = conn.Conn
	reconn := p.Connections[conn.ID]
	fmt.Println(p.Connections[conn.ID].Conn)

	workers.GlobalPool.QueueJob(NewWriteJob(reconn, []byte("Reconnecting ...\n"), 0))

	logger.Default.Info().Msgf("Notifications missed: %+v", reconn.Notify)
	toSend := make([]string, 0, len(reconn.Notify))
	toSend = append(toSend, reconn.Notify...)
	for _, msg := range toSend {
		workers.GlobalPool.QueueJob(NewWriteJob(reconn, []byte(msg), 1))
	}
	reconn.Notify = make([]string, 0)

	for len(conn.stop) > 0 {
		<-conn.stop
		fmt.Println("##### INTERRUPTED A STOP SIGNAL ####")
	}

	return p.Connections[conn.ID]
}

func (p *ConnPool) GetConn(id string) *Connection {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.Connections[id]
}

func (p *ConnPool) CloseConn(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.Connections[id].Close()
}
