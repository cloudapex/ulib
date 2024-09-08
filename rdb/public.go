package rdb

/******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2014-10-18
Description: Redis控制器(单例)
******************************************************************************/
import (
	"fmt"

	"github.com/cloudapex/ulib/ctl"
	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
)

var (
	Ctl IContrler // 默认Cache控制器

	mapCoders = make(map[ECoding]Coder)
)

// 安装控制器
func Install(confs []*Config) ctl.IControler {
	c := ctl.Install(Controller(confs))
	util.Cast(Ctl == nil, func() { Ctl = c.(IContrler) }, nil)
	return Ctl
}

// Pool 获取指定pool
func Pool(dbName string) IPooler { return Ctl.Use(dbName) }

// Connector 获取指定pool的连接 !!!注意释放!!!
func Connector(dbName string) redis.Conn {
	p := Ctl.Use(dbName)
	if p == nil {
		log.Error("rdb.pool[%q] not found", dbName)
		return nil
	}
	return p.Get()
}

// RegistCoder 注册编解码器
func RegistCoder(name ECoding, coder Coder) {
	mapCoders[name] = coder
}

// Pipe Pipelining管道批量并发操作,统一返回结果 (非原子)
func Pipe(sends ...sender) error {
	if len(sends) == 0 {
		return nil
	}
	// only same db index
	name := sends[0].K.DB
	for _, it := range sends {
		if it.K.DB != name {
			return fmt.Errorf("db name of sends must be samed")
		}
	}

	c := Connector(name)
	defer c.Close()

	// sends
	for _, it := range sends {
		it.K.send = &sendcc{c, 0}
		it.Sends()
	}

	// flush
	if err := c.Flush(); err != nil {
		return err
	}

	// replys
	for _, it := range sends {
		rps := []*reply{}
		for i := 0; i < it.K.send.count; i++ {
			r, err := c.Receive()
			rps = append(rps, &reply{rep: r, err: err, coding: it.K.Coding})
		}
		if it.Reply != nil {
			it.Reply(rps)
		}
	}

	// clear
	for _, it := range sends {
		it.K.send = nil
	}
	return nil
}

// Exec 事务原子操作,统一返回结果 ( sender.Reply 赋值不起作用)
func Exec(sends ...sender) *reply {
	if len(sends) == 0 {
		return &reply{err: redis.ErrNil}
	}
	// only same db index
	name := sends[0].K.DB
	for _, it := range sends {
		if it.K.DB != name {
			return &reply{err: fmt.Errorf("db name of sends must be same, %s !=%s", it.K.DB, name)}
		}
	}

	c := Connector(name)
	defer c.Close()

	// sends
	c.Send("MULTI")
	for _, it := range sends {
		it.K.send = &sendcc{c, 0}
		it.Sends()
	}

	r, err := c.Do("EXEC")

	// clear
	for _, it := range sends {
		it.K.send = nil
	}
	return &reply{rep: r, err: err}
}

// Command 直接执行命令
func Command(dbName string, coding ECoding, command string, args ...interface{}) *reply {
	return (&Key{DB: dbName, K: "", Coding: coding}).do(command, args...)
}
