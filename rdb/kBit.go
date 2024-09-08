package rdb

import "github.com/gomodule/redigo/redis"

type Bit struct{ Key }

// SetBit 特定位设置 0 or 1
func (k *Bit) SetBit(offset, val_0_or_1 int64) *reply {
	return k.do("setbit", k.K, offset, val_0_or_1)
}

// GetBit 特定位取值 0 or 1
func (k *Bit) GetBit(offset int64) *reply {
	return k.do("getbit", k.K, offset)
}

// BitCount 被设置为 1 的位的数量
func (k *Bit) BitCount(beginEnd ...int64) *reply {
	if len(beginEnd) == 2 {
		return k.do("bitcount", k.K, beginEnd[0], beginEnd[1])
	}
	return k.do("bitcount", k.K)
}

// opt 包含 and、or、xor、not
func (k *Bit) BitOpt(opt, destKey string, keys ...string) *reply {
	return k.do("bitop", opt, destKey, redis.Args{}.Add(keys).AddFlat(keys))
}
