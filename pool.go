package pool

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
)

type factory func() (Poolable, error)

type GenericPool struct {
	sync.Mutex
	pool        chan Poolable
	maxOpen     int  // 池中最大资源数
	numOpen     int  // 当前池中资源数
	minOpen     int  // 池中最少资源数
	closed      bool // 池是否已关闭
	maxLifetime time.Duration
	factory     factory // 创建连接的方法
}

func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}

	p := &GenericPool{
		maxOpen:     maxOpen,
		minOpen:     minOpen,
		maxLifetime: maxLifetime,
		factory:     factory,
		pool:        make(chan Poolable, maxOpen),
	}

	for i := 0; i < minOpen; i++ {
		closer, err := factory()
		if err != nil {
			continue
		}
		p.numOpen++
		p.pool <- closer
	}

	return p, nil
}

func (p *GenericPool) Get() (Poolable, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}

	for {
		closer, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}

		// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
		if p.maxLifetime > 0 && closer.ActiveTime().Add(p.maxLifetime).Before(time.Now()) {
			p.Close(closer)
			continue
		}

		return closer, nil
	}
}

func (p *GenericPool) getOrCreate() (Poolable, error) {
	// 检查池中是否有对象
	select {
	case closer := <-p.pool:
		return closer, nil
	default:
	}

	p.Lock()

	if p.numOpen >= p.maxOpen {
		closer := <-p.pool
		p.Unlock()

		return closer, nil
	}

	// 新建连接
	closer, err := p.factory()
	if err != nil {
		p.Unlock()
		return nil, err
	}
	p.numOpen++
	p.Unlock()

	return closer, nil
}

// 释放单个资源到连接池
func (p *GenericPool) Put(closer Poolable) error {
	if p.closed {
		return ErrPoolClosed
	}

	// 无需加锁防止死锁
	// 加锁
	// p.Lock()

	// select-case
	// 控制锁释放
	select {
	case p.pool <- closer:
	default:
	}

	// 释放锁
	// p.Unlock()

	return nil
}

// 关闭单个资源
func (p *GenericPool) Close(closer Poolable) error {
	p.Lock()
	closer.Close()
	p.numOpen--
	p.Unlock()
	return nil
}

// 关闭连接池，释放所有资源
func (p *GenericPool) Shutdown() error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	close(p.pool)
	for closer := range p.pool {
		closer.Close()
		p.numOpen--
	}
	p.closed = true
	p.Unlock()
	return nil
}
