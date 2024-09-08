package rdb

import (
	"github.com/cloudapex/ulib/ctl"

	"github.com/gomodule/redigo/redis"
)

// 控制器接口
type IContrler interface {
	ctl.IControler

	// 获取某个连接池对象
	Use(name string) IPooler
}

// 连接池接口
type IPooler interface {
	Mode() EMode
	Get() redis.Conn
	Close() error
}

// 所有Key接口
type IKeyer interface {
	key() *Key
}

// > 编码器接口
type Coder struct {
	Encoder func(v interface{}) ([]byte, error)
	Decoder func(data []byte, v interface{}) error
}

// > 互斥锁接口
type IMutexer interface {
	Lock() error
	Unlock() (bool, error)
	Extend() (bool, error)
}
