package pool

import (
	"fmt"
	// "io"
	"log"
	"strconv"
	"testing"
	"time"
)

// 定义被管理的对象
type CloserObj struct {
	Name     string
	activeAt time.Time
}

func (p *CloserObj) Close() error {
	fmt.Println(p.Name, "closed")
	return nil
}

func (p *CloserObj) ActiveTime() time.Time {
	return p.activeAt
}

// 测试实例化池
func TestNewGenericPool(t *testing.T) {
	_, err := NewGenericPool(0, 10, time.Minute*10, func() (Poolable, error) {
		time.Sleep(time.Second)
		return &CloserObj{Name: "test", activeAt: time.Now()}, nil
	})
	if err != nil {
		t.Error(err)
	}
}

// 测试存取对象
func TestPoolGet(t *testing.T) {
	pool, err := NewGenericPool(0, 5, time.Minute*10, func() (Poolable, error) {
		time.Sleep(time.Second)
		name := strconv.FormatInt(time.Now().Unix(), 10)
		log.Printf("%s created", name)
		// TODO: FIXME &DemoCloser{Name: name}后，pool.Acquire陷入死循环
		return &CloserObj{Name: name, activeAt: time.Now()}, nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 100; i++ {
		s, err := pool.Get()
		if err != nil {
			t.Error(err)
			return
		}
		pool.Close(s)
		// pool.Put(s)
	}
}

// 测试关闭资源
func TestClose(t *testing.T) {
	pool, err := NewGenericPool(0, 5, time.Minute*10, func() (Poolable, error) {
		time.Sleep(time.Second)
		name := strconv.FormatInt(time.Now().Unix(), 10)
		log.Printf("%s created", name)
		// TODO: FIXME &DemoCloser{Name: name}后，pool.Acquire陷入死循环
		return &CloserObj{Name: name, activeAt: time.Now()}, nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 10; i++ {
		s, err := pool.Get()
		if err != nil {
			t.Error(err)
			return
		}
		pool.Close(s)
		// pool.Put(s)
	}
}

// 测试关闭池
func TestShutdown(t *testing.T) {
	pool, err := NewGenericPool(0, 10, time.Minute*10, func() (Poolable, error) {
		time.Sleep(time.Second)
		return &CloserObj{Name: "test"}, nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	if err := pool.Shutdown(); err != nil {
		t.Error(err)
		return
	}
	if _, err := pool.Get(); err != ErrPoolClosed {
		t.Error(err)
	}
}
