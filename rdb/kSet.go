package rdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-05-13
Description: Set数据结构
*******************************************************************************/
import "github.com/gomodule/redigo/redis"

type Set struct{ Key }

// Set.Add 添加元素
func (k *Set) Add(members ...interface{}) *reply {
	return k.do("sadd", redis.Args{}.Add(k.K).AddFlat(members)...)
}

// Set.Count 集合元素个数
func (k *Set) Count() *reply {
	return k.do("scard", k.K)
}

// Set.Members 返回集合中成员
func (k *Set) Members() *reply {
	return k.do("smembers", k.K)
}

// Set.IsMember 判断元素是否是集合成员
func (k *Set) IsMember(member interface{}) *reply {
	return k.do("sismember", k.K, member)
}

// Set.Pop 返回并移除一个元素
func (k *Set) Pop() *reply {
	return k.do("spop", k.K)
}

// Set.RandMember 随机返回一个或多个元素
func (k *Set) RandMember(count ...int64) *reply {
	if len(count) > 0 {
		return k.do("srandmember", k.K, count[0])
	}
	return k.do("srandmember", k.K)
}

// Set.Rem 移除指定的元素
func (k *Set) Rem(members ...interface{}) *reply {
	return k.do("srem", redis.Args{}.Add(k.K).AddFlat(members)...)
}

// Set.Move 将元素从集合移至另一个集合
func (k *Set) Move(sourceKey, destinationKey string, member interface{}) *reply {
	return k.do("smove", sourceKey, destinationKey, member)
}

// Set.Diff 返回一或多个集合的差集
func (k *Set) Diff(keys []string) *reply {
	return k.do("sdiff", redis.Args{}.AddFlat(keys)...)
}

// Set.DiffStore 将一或多个集合的差集保存至另一集合(destinationKey)
func (k *Set) DiffStore(destinationKey string, keys []string) *reply {
	return k.do("sdiffstore", redis.Args{}.Add(destinationKey).AddFlat(keys)...)
}

// Set.InterStore 将keys的集合的并集 写入到 destinationKey中
func (k *Set) InterStore(destinationKey string, keys []string) *reply {
	return k.do("sinterstore", redis.Args{}.Add(destinationKey).AddFlat(keys)...)
}

// Set.Inter 一个或多个集合的交集
func (k *Set) Inter(keys []string) *reply {
	return k.do("sinter", redis.Args{}.AddFlat(keys)...)
}

// Set.Union 返回集合的并集
func (k *Set) Union(keys []string) *reply {
	return k.do("sunion", redis.Args{}.AddFlat(keys)...)
}

// Set.UnionStore 将 keys 的集合的并集 写入到 destinationKey 中
func (k *Set) UnionStore(destinationKey string, keys []string) *reply {
	return k.do("sunionstore", redis.Args{}.Add(destinationKey).AddFlat(keys)...)
}
