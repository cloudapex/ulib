package rdb

import "github.com/cloudapex/ulib/util"

type List struct {
	Key
	FixSize int // 定长队列(>0时有效)
}

// List.LPush 向列表头插入元素 (whenListExist 当列表存在时才添加)
func (k *List) LPush(value interface{}, whenListExist ...bool) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "List.LPush")
	}
	value = v
	if util.DefaultVal(whenListExist) {
		return k.do("lpushx", k.K, value)
	}
	return k.do("lpush", k.K, value)
}

// List.RPush 将指定元素插入列表末尾 (whenListExist 当列表存在时才添加)
func (k *List) RPush(value interface{}, whenListExist ...bool) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "List.RPush")
	}
	value = v
	if k.FixSize > 0 { // 忽略 whenListExist 参数
		return k.do("eval", lua_fix_size_rpush, 1, k.K, value, k.FixSize)
	}
	if util.DefaultVal(whenListExist) {
		return k.do("rpushx", k.K, value)
	}
	return k.do("rpush", k.K, value)
}

// List.LInsert 将元素插入指定位置 position:BEFORE|AFTER,当 pivot 不存在于列表 key 时，不执行任何操作。当 key 不存在时， key 被视为空列表，不执行任何操作。
func (k *List) LInsert(position, pivot, value string) *reply {
	return k.do("linsert", k.K, position, pivot, value)
}

// List.LPop 移除并返回列表头元素
func (k *List) LPop() *reply {
	return k.do("lpop", k.K)
}

// List.BLpop 阻塞并弹出头元素
func (k *List) BLpop(timeout interface{}) *reply {
	return k.do("blpop", k.K, timeout)
}

// List.RPop 移除并返回列表尾元素
func (k *List) RPop() *reply {
	return k.do("cpop", k.K)
}

// List.BRpop 阻塞并弹出末尾元素
func (k *List) BRpop(timeout interface{}) *reply {
	return k.do("brpop", k.K, timeout)
}

// List.LIndex 返回指定位置的元素
func (k *List) LIndex(index int) *reply {
	return k.do("lindex", k.K, index)
}

// List.LRange 获取指定区间的元素(全部:(0,-1))
func (k *List) LRange(start, stop interface{}) *reply {
	return k.do("lrange", k.K, start, stop)
}

// List.LSet 设置指定位元素
func (k *List) LSet(index, value interface{}) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "List.LSet")
	}
	value = v
	return k.do("lset", k.K, index, value)
}

// List.RPoplpush 弹出source尾元素并返回，将弹出元素插入destination列表的开头
func (k *List) RPoplpush(key, source, destination string) *reply {
	return k.do("rpoplpush", key, source, destination)
}

// List.BRpoplpush 阻塞并弹出尾元素，将弹出元素插入另一列表的开头
func (k *List) BRpoplpush(key, source, destination string, timeout interface{}) *reply {
	return k.do("brpoplpush", key, source, destination, timeout)
}

// List.LRem 移除元素,count = 0 : 移除表中所有与 value 相等的值,count!=0,移除与 value 相等的元素，数量为 count的绝对值
func (k *List) LRem(count, value interface{}) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "List.LRem")
	}
	value = v
	return k.do("lrem", k.K, count, value)
}

// List.LTrim 列表裁剪，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。-1 表示尾部
func (k *List) LTrim(start, stop interface{}) *reply {
	v, err := Encode(k.Coding, stop)
	if err != nil {
		return Reply(nil, err, k.Coding, "List.LTrim")
	}
	stop = v
	return k.do("ltrim", k.K, start, stop)
}

// ------------------------------------------------

// LPush fix_size queue script
var lua_fix_size_rpush = `
	local size = redis.call('llen',KEYS[1]);
	if tonumber(size) >= tonumber(ARGV[2]) then
		redis.call('lpop', KEYS[1])
	end
	redis.call('rpush',KEYS[1],ARGV[1])
	local result = redis.call('llen',KEYS[1])
	return result
`
