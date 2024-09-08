package rdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-05-13
Description: 相关api封装Key
*******************************************************************************/
import (
	"fmt"
	"time"

	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
)

type Key struct {
	DB     string        // 哪个db中(表示哪个db索引)
	K      string        // Key的名称(建议格式:'basexxx:param1=%d,param2=%s,....')
	Coding ECoding       // 编码模式
	Ttl    time.Duration // 存活时间ms(仅作存储,需要自行调用k.Overdue())

	send *sendcc // 用于Send的连接
}

// key.Valid 键是否有效
func (k *Key) Valid() bool { return k.DB != "" && k.K != "" }

// key.Keys 查找键 [*模糊查找]
func (k *Key) Keys() *reply {
	return k.do("keys", k.K)
}

// key.Existed 判断key是否存在
func (k *Key) Existed() *reply {
	return k.do("exists", k.K)
}

// key.RandomKey 随机返回一个key
func (k *Key) RandomKey() *reply {
	return k.do("randomkey")
}

// key.Type 返回值的数据类型
func (k *Key) Type() *reply {
	return k.do("type", k.K)
}

// key.Delete 删除key(默认删自己,否则删入参指定的keys)
func (k *Key) Delete(keys ...string) *reply {
	if len(keys) == 0 {
		return k.do("del", k.K)
	}
	return k.do("del", redis.Args{}.AddFlat(keys)...)
}

// key.Del 重命名
func (k *Key) Rename(newKey string) *reply {
	return k.do("rename", k.K, newKey)
}

// key.RenameNX 仅当newkey不存在时重命名
func (k *Key) RenameNX(newKey string) *reply {
	return k.do("renamenx", k.K, newKey)
}

// key.RenameNX 序列化key
func (k *Key) Dump() *reply {
	return k.do("dump", k.K)
}

// key.Restore 反序列化
func (k *Key) Restore(ttlMS, serializedValue interface{}) *reply {
	return k.do("restore", k.K, ttlMS, serializedValue)
}

// key.Expire 设置Key过期时间段(秒)
func (k *Key) Expire(seconds int64) *reply {
	return k.do("expire", k.K, seconds)
}

// key.PExpire 设置Key过期时间段(毫秒)
func (k *Key) PExpire(millisecond int64) *reply {
	return k.do("pexpire", k.K, millisecond)
}

// key.ExpireAt 设置Key过期时间点(秒)
func (k *Key) ExpireAt(timestamp int64) *reply {
	return k.do("expireat", k.K, timestamp)
}

// key.PExpireAt 设置Key过期时间点(毫秒)
func (k *Key) PExpireAt(timestamp_ms int64) *reply {
	return k.do("pexpireat", k.K, timestamp_ms)
}

// key.Refresh 刷新Key过期时间 by Ttl
func (k *Key) Refresh(ttl ...time.Duration) *reply {
	ttl_ms := k.Ttl.Milliseconds()
	util.Cast(len(ttl) > 0, func() { ttl_ms = ttl[0].Milliseconds() }, nil)

	if ttl_ms == 0 {
		return Reply(nil, ErrInvalidTTL, ECod_None, fmt.Sprintf("%v %s %d", "pexpireat", k.K, 0))
	}
	return k.PExpire(ttl_ms)
}

// key.Persist 移除Key的过期设置
func (k *Key) Persist() *reply {
	return k.do("persist", k.K)
}

// key.TTL 剩余过期时间(秒)
func (k *Key) TTL() *reply {
	return k.do("ttl", k.K)
}

// key.PTTL 剩余过期时间(毫秒)
func (k *Key) PTTL() *reply {
	return k.do("pttl", k.K)
}

// key.Select 选择库
func (k *Key) Select(db int64) *reply {
	return k.do("select", db)
}

// key.Move 同实例不同库间的键移动
func (k *Key) Move(db int64) *reply {
	return k.do("move", k.K, db)
}

// Clusted 是否是集群模式
func (k *Key) Clusted() bool { return Pool(k.DB).Mode() == EMode_Cluster }

// --------------- internal

func (k *Key) key() *Key { return k }

func (k *Key) do(command string, args ...interface{}) *reply {
	if k.send != nil {
		k.send.count++
		k.send.conn.Send(command, args...)
		return nil
	}

	c := Connector(k.DB)
	defer c.Close()

	r, err := c.Do(command, args...)
	return Reply(r, err, k.Coding, fmt.Sprintf("k:%s command:%v arg:%v", k.K, command, args))
}
