package rdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-05-13
Description: Hash数据结构
*******************************************************************************/
import (
	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
)

type Hash struct{ Key }

// Hash.Set 为字段设置其值(noExist为true 表示字段不存则设置其值)
func (k *Hash) Set(field, value interface{}, noExist ...bool) *reply {
	v, err := Encode(k.Coding, value)
	if err != nil {
		return Reply(nil, err, k.Coding, "Hash.Set")
	}
	value = v

	if util.DefaultVal(noExist) {
		return k.do("hsetnx", k.K, field, value)
	}
	return k.do("hset", k.K, field, value)
}

// Hash.Get 获取指定字段值
func (k *Hash) Get(field interface{}) *reply {
	return k.do("hget", k.K, field)
}

// Hash.Getm 返回多个字段的值(reply.Xxxs)
func (k *Hash) Getm(fields ...interface{}) *reply {
	return k.do("hmget", redis.Args{}.Add(k.K).AddFlat(fields)...)
}

// Hash.SetMap 设置多个字段及值(map[field]value)
func (k *Hash) SetMap(mp map[interface{}]interface{}) *reply {
	for _k, _v := range mp {
		v, err := Encode(k.Coding, _v)
		if err != nil {
			return Reply(nil, err, k.Coding, "Hash.SetMap")
		}
		mp[_k] = v
	}
	return k.do("hmset", redis.Args{}.Add(k.K).AddFlat(mp)...)
}

// Hash.GetMap 返回多个字段以及值(reply.HashMap)
func (k *Hash) GetMap(fields ...interface{}) *reply {
	r := k.do("hmget", redis.Args{}.Add(k.K).AddFlat(fields)...)
	r.ext = fields
	return r
}

// Hash.GetAll 设置对象的字段及值(reply.XxxMap)
func (k *Hash) GetAll() *reply {
	return k.do("hgetall", k.K)
}

// Hash.SetStruct 设置对象的字段及值(*struct)
func (k *Hash) SetStruct(obj interface{}) *reply {
	return k.do("hmset", redis.Args{}.Add(k.K).AddFlat(obj)...)
}

// Hash.GetStruct 获取对象的字段及值(reply.ScanStruct)
func (k *Hash) GetStruct() *reply {
	return k.do("hgetall", k.K)
}

// Hash.Del 字段删除()
func (k *Hash) Del(fields ...interface{}) *reply {
	return k.do("hdel", redis.Args{}.Add(k.K).AddFlat(fields)...)
}

// Hash.Exists 判断字段是否存在
func (k *Hash) Exists(field interface{}) *reply {
	return k.do("hexists", k.K, field)
}

// Hash.Len 返回字段数量
func (k *Hash) Len() *reply {
	return k.do("hlen", k.K)
}

// Hash.Fields 返回所有字段
func (k *Hash) Fields() *reply {
	return k.do("hkeys", k.K)
}

// Hash.Values 返回所有字段的值
func (k *Hash) Values() *reply {
	return k.do("hvals", k.K)
}

// Hash.IncrBy 为指定字段值增加
func (k *Hash) IncrBy(field interface{}, increment interface{}) *reply {
	return k.do("hincrby", k.K, field, increment)
}

// Hash.IncrByFloat 为指定字段值增加浮点数
func (k *Hash) IncrByFloat(field interface{}, increment float64) *reply {
	return k.do("hincrbyfloat", k.K, field, increment)
}
