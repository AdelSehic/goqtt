package server

import "sync"

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

func (p *ConnPool) AddConn(conn *Connection) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.Connections[conn.ID] = conn
}

func (p *ConnPool) GetConn(id string) *Connection {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.Connections[id]
}
