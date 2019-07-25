/*
@Time : 2019-07-24 15:25
@Author : zr
*/
package connpool

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrTimeOut     = errors.New("timeout")
	ErrCapZero     = errors.New("cap不能小于0")
	ErrTimeOutZero = errors.New("connTimeOut不能小于0")
)

type Connection interface {
	Close() error
	IsOpen() bool
}

type Factory func() (Connection, error)

type Pool struct {
	sync.Mutex
	numOpen     int
	conns       chan Connection
	factory     Factory
	connTimeOut time.Duration
}

func NewPool(factory Factory, cap int, connTimeOut time.Duration) (*Pool, error) {

	if cap <= 0 {
		return nil, ErrCapZero
	}
	if connTimeOut <= 0 {
		return nil, ErrTimeOutZero
	}

	pool := &Pool{
		conns:       make(chan Connection, cap),
		factory:     factory,
		connTimeOut: connTimeOut,
	}

	return pool, nil
}
func (p *Pool) newConn() (conn Connection, err error) {
	conn, err = p.factory()
	return
}

func (p *Pool) Get() (conn Connection, err error) {
	// 连接池未满
	p.Lock()
	if p.numOpen < cap(p.conns) {

		conn, err = p.newConn()
		if err != nil {
			p.Unlock()
			return
		}
		p.numOpen++
		p.Unlock()
		return
	}
	p.Unlock()

	select {
	case conn, err = <-p.conns:
		{
			if err != nil {
				return nil, ErrTimeOut
			}
		}
	case <-time.After(p.connTimeOut):
		{
			return nil, ErrTimeOut
		}
	}
	return
}

func (p *Pool) Put(conn Connection) {
	p.Lock()
	if !conn.IsOpen() {
		p.numOpen--
		p.Unlock()
		return
	}
	p.Unlock()

	select {
	case p.conns <- conn:
		{
		}
	default:
	_:
		conn.Close()
	}
}
