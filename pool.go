/*
@Time : 2019-07-24 15:25
@Author : zr
*/
package connpool

import (
	"errors"
	"time"
)

var ErrTimeOut = errors.New("timeout")

type ConnRes interface {
	Close() error
}
type Factory func() (ConnRes, error)

type Pool struct {
	conns       chan ConnRes
	factory     Factory
	connTimeOut time.Duration
}

func NewPool(factory Factory, cap int, connTimeOut time.Duration) (*Pool, error) {
	if cap <= 0 {
		return nil, errors.New("cap不能小于0")
	}
	if connTimeOut <= 0 {
		return nil, errors.New("connTimeOut不能小于0")
	}

	pool := &Pool{
		conns:       make(chan ConnRes, cap),
		factory:     factory,
		connTimeOut: connTimeOut,
	}

	for i := 0; i < cap; i++ {
		//通过工厂方法创建连接资源
		connRes, err := pool.new()
		if err != nil {
			return nil, err
		}
		//将连接资源插入通道中
		pool.conns <- connRes
	}

	return pool, nil
}
func (p *Pool) new() (ConnRes, error) {
	conn, err := p.factory()
	return conn, err
}

func (p *Pool) Get() (conn ConnRes, err error) {

	select {
	case conn = <-p.conns:
		{
		}
	case <-time.After(p.connTimeOut):
		{
			return nil, ErrTimeOut
		}
	}
	return
}

func (p *Pool) Put(conn ConnRes) {
	select {
	case p.conns <- conn:
		{
		}
	default:
	_:
		conn.Close()
	}
}
