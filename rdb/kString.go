package rdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-05-13
Description: String 数据结构
*******************************************************************************/

import (
	"fmt"
	"time"

	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
)

type String struct{ Key }

// String.Set 设置值(会更新ttl 优先使用 expire)
func (k *String) Set(mode ESetMode, value interface{}, expire ...time.Duration) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "String.Set")
	}
	value = v
	util.Cast(len(expire) == 0 && k.Ttl != 0, func() { expire = append(expire, k.Ttl) }, nil)

	switch mode {
	case ESet_WhenExist:
		if len(expire) == 0 || expire[0] == 0 {
			return k.do("set", k.K, value, "XX")
		}
		return k.do("set", k.K, value, "PX", expire[0].Milliseconds(), "XX")
	case ESet_WhenNoExist:
		if len(expire) == 0 || expire[0] == 0 {
			return k.do("set", k.K, value, "NX")
		}
		return k.do("set", k.K, value, "PX", expire[0].Milliseconds(), "NX")
	}

	if len(expire) == 0 || expire[0] == 0 {
		return k.do("set", k.K, value)
	}
	return k.do("set", k.K, value, "PX", expire[0].Milliseconds())
}

// String.Upd 更新值(不会更新expire时间)
func (k *String) Upd(mode ESetMode, value interface{}) *reply {
	return k.Set(mode, value, 0)
}

// String.Setm 设置多个值(不更新ttl)
func (k *String) Setm(kv map[string]interface{}) *reply {
	for _k, _v := range kv {
		v, err := Encode(k.Coding, _v)
		if err != nil {
			return Reply(nil, err, k.Coding, "String.Setm")
		}
		kv[_k] = v
	}
	return k.do("mset", redis.Args{}.AddFlat(kv))
}

// String.Setms 设置多个值(不更新ttl)(以k.K作为fmt模块)
func (k *String) Setms(args [][]interface{}, vals []interface{}) *reply {
	_vals := make([]string, len(vals))
	for n, _v := range vals {
		v, err := Encode(k.Coding, _v)
		if err != nil {
			return Reply(nil, err, k.Coding, "String.Setms")
		}
		_vals[n] = fmt.Sprintf("%v", v)
	}
	kvs := []string{}
	for n, _args := range args {
		kvs = append(kvs, fmt.Sprintf(k.K, _args...), _vals[n])
	}
	return k.do("mset", redis.Args{}.AddFlat(kvs))
}

// String.SetmNx key不存在时设置多个值(当且仅当所有给定键都不存在时,为所有给定键设置值)
func (k *String) SetmNx(kv map[string]interface{}) *reply {
	for _k, _v := range kv {
		v, err := Encode(k.Coding, _v)
		if err != nil {
			return Reply(nil, err, k.Coding, "String.SetmNx")
		}
		kv[_k] = v
	}
	return k.do("msetnx", redis.Args{}.AddFlat(kv))
}

// String.SetRange 设置子字符串
func (k *String) SetRange(value interface{}, offset int64) *reply {
	return k.do("setrange", k.K, offset, value)
}

// String.Get 获取值
func (k *String) Get() *reply {
	return k.do("get", k.K)
}

// String.GetSet 设置并返回旧值
func (k *String) GetSet(value interface{}) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "String.GetSet")
	}
	value = v
	return k.do("getset", k.K, value)
}

// String.GetRange 获取子字符串
func (k *String) GetRange(start, end int64) *reply {
	return k.do("getrange", k.K, start, end)
}

// String.Getm 返回多个key的值
func (k *String) Getm(keys []string) *reply {
	return k.do("mget", redis.Args{}.AddFlat(keys)...)
}

// String.Getms 返回多个key的值(以k.K作为fmt模块)
func (k *String) Getms(args [][]interface{}) *reply {
	keys_ := []string{}
	for _, _args := range args {
		keys_ = append(keys_, fmt.Sprintf(k.K, _args...))
	}
	return k.do("mget", redis.Args{}.AddFlat(keys_)...)
}

// String.Append 旧值的尾部追加值。
func (k *String) Append(value interface{}) *reply {
	return k.do("append", k.K, value)
}

// String.Incr 自增(并返回加之后的值)
func (k *String) Incr() *reply {
	return k.do("incr", k.K)
}

// String.IncrBy 增加指定值(并返回增加之后的值)
func (k *String) IncrBy(increment int64) *reply {
	return k.do("incrby", k.K, increment)
}

// String.IncrByFloat 增加一个浮点值(并返回增加之后的值)
func (k *String) IncrByFloat(increment float64) *reply {
	return k.do("incrbyfloat", k.K, increment)
}

// String.Decr 自减(并返回减之后的值)
func (k *String) Decr() *reply {
	return k.do("decr", k.K)
}

// String.DecrBy 自减指定值(并返回减之后的值)
func (k *String) DecrBy(increment int64) *reply {
	return k.do("decrby", k.K, increment)
}

// String.LuaIncrBy 安全增减键值(不会出现负值), Result.Reply==nil表示失败, Result.Int()是最新值
func (k *String) LuaIncrBy(amount int) *reply {
	c := Connector(k.DB)
	defer c.Close()
	r, err := lua_incrby.Do(c, k.K, amount)
	return Reply(r, err, k.Coding, "String.LuaIncrBy")
}

// ------------------------------------------------

// Only increase the value if the key exists.
var lua_incrby = redis.NewScript(1, `
	local key = KEYS[1];
	local amount = tonumber(ARGV[1]);
	if amount < 0 then
		local n = redis.call("GET", key);
		n = tonumber(n or 0);
		if n < -amount then
			return nil;
		else
			n = redis.call("INCRBY", key, amount);
			return n;
		end
	elseif redis.call("EXISTS", key) == 1 then
		local n = redis.call("INCRBY", key, amount);
		return n;
	end
	return nil;
`)
