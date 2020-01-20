package pool

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 定义被管理的对象
type CloserObj struct {
	Name     string
	activeAt time.Time
}

// 全局控制goroutine
var waitgroup sync.WaitGroup
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
	pool, err := NewGenericPool(0, 20, time.Minute*10, func() (Poolable, error) {
		time.Sleep(20 * time.Millisecond)
		name := "test-" + strconv.FormatInt(time.Now().Unix(), 10) + randSeq(6)
		log.Printf("%s created", name)
		// TODO: FIXME &DemoCloser{Name: name}后，pool.Acquire陷入死循环
		return &CloserObj{Name: name, activeAt: time.Now()}, nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 200; i++ {
		// 增加一个
		waitgroup.Add(1)
		// time.Sleep(50 * time.Millisecond)
		// 启动goroutine
		go func() {
			// 获取资源
			s, err := pool.Get()
			// 休眠1秒等待其他goroutine获取
			time.Sleep(10 * time.Millisecond)
			if err != nil {
				t.Error(err)
				return
			}
			// 回放资源
			pool.Put(s)
			// 释放一个
			waitgroup.Done()
		}()
	}
	// 等待所有goroutine结束
	waitgroup.Wait()
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

// 生成指定长度的串
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
