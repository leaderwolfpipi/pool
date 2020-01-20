package pool

import (
	"io"
	"time"
)

// 通用池接口
type Pool interface {
	Get() (Poolable, error) // 获取资源
	Put(Poolable) error     // 释放资源
	Close(Poolable) error   // 关闭资源
	Shutdown() error        // 关闭池
}

// 可池化接口
// 用于描述被管理的对象
type Poolable interface {
	io.Closer
	ActiveTime() time.Time
}
