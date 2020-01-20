# 通用池化

## 导入包
`import "github.com/leaderwolfpipi/pool"`

## 使用案例
```go
package main

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

// 存取对象
func main() {
	pool, err := NewGenericPool(0, 5, time.Minute*10, func() (Poolable, error) {
		time.Sleep(time.Second)
		name := strconv.FormatInt(time.Now().Unix(), 10)
		log.Printf("%s created", name)
		// TODO: FIXME &DemoCloser{Name: name}后，pool.Acquire陷入死循环
		return &CloserObj{Name: name, activeAt: time.Now()}, nil
	})
	if err != nil {
		fmt.Error(err)
		return
	}
	for i := 0; i < 100; i++ {
		s, err := pool.Get()
		if err != nil {
			fmt.Error(err)
			return
		}
		pool.Close(s)
		// pool.Put(s)
	}
}

```