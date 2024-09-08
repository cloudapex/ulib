package rdb

import (
	"fmt"

	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/gomodule/redigo/redis"
)

type Zset struct{ Key }

// Zset.Add 添加|更新元素
func (k *Zset) Add(mode ESetMode, member, score interface{}) *reply {
	switch mode {
	case ESet_WhenExist:
		return k.do("zadd", k.K, "XX", score, member)
	case ESet_WhenNoExist:
		return k.do("zadd", k.K, "NX", score, member)
	case ESet_WhenValueLT:
		return k.do("zadd", k.K, "LT", score, member)
	case ESet_WhenValueGT:
		return k.do("zadd", k.K, "GT", score, member)
	}
	return k.do("zadd", k.K, score, member)
}

// Zset.AddArr 添加|更新多元素 member1 score1, ... memberN,scoreN
func (k *Zset) AddArr(ar ...interface{}) *reply {
	if n := len(ar); n == 0 || n%2 != 0 {
		return Reply(nil, ErrZsetAddArrInvalid, ECod_None, fmt.Sprintf("%v %s %d", "zadd", k.K, 0))
	}
	for i := 0; i < len(ar); i += 2 {
		ar[i], ar[i+1] = ar[i+1], ar[i]
	}
	return k.do("zadd", redis.Args{}.Add(k.K).AddFlat(ar)...)
}

// Zset.AddMap 添加|更新多元素 map[member]score
func (k *Zset) AddMap(mp map[interface{}]interface{}) *reply {
	ar := []interface{}{}
	for k, v := range mp {
		ar = append(ar, v, k)
	}
	return k.do("zadd", redis.Args{}.Add(k.K).AddFlat(ar)...)
}

// Zset.Incr 增加元素分数
func (k *Zset) Incr(member, increment interface{}) *reply {
	return k.do("zincrby", k.K, increment, member)
}

// Zset.Card 返回元素数量
func (k *Zset) Card() *reply {
	return k.do("zcard", k.K)
}

// Zset.Rank 返回指定元素的排名
func (k *Zset) Rank(e ESort, member interface{}) *reply {
	if e == ESort_Desc {
		return k.do("zrevrank", k.K, member)
	}
	return k.do("zrank", k.K, member)
}

// Zset.Topn 返回Top排行元素
func (k *Zset) Topn(e ESort, start, stopN int, withScore ...bool) *reply {
	stopN = mathutil.Max(stopN, 1)
	if e == ESort_Desc {
		return k.Revrange(start, stopN-1, withScore...)
	}
	return k.Range(start, stopN-1, withScore...)
}

// Zset.Topn 返回Top排行元素
func (k *Zset) TopnByScore(e ESort, min, max interface{}, withScore ...bool) *reply {
	if e == ESort_Desc {
		return k.Revrangebyscore(min, max, withScore...)
	}
	return k.Rangebyscore(min, max, withScore...)
}

// Zset.Score 返回指定元素的分数
func (k *Zset) Score(member interface{}) *reply {
	return k.do("zscore", k.K, member)
}

// Zset.Count 返回集合两个分数间的元素数(min<=score<=max)
func (k *Zset) Count(min, max interface{}) *reply {
	return k.do("zcount", k.K, min, max)
}

// Zset.Range 返回指定区间内的元素([start, stop])
func (k *Zset) Range(start, stop interface{}, withScore ...bool) *reply {
	if len(withScore) > 0 && withScore[0] {
		return k.do("zrange", k.K, start, stop, "WITHSCORES")
	}
	return k.do("zrange", k.K, start, stop)
}

// Zset.Revrange 倒序返回指定区间内的元素([start, stop])
func (k *Zset) Revrange(start, stop interface{}, withScore ...bool) *reply {
	if len(withScore) > 0 && withScore[0] {
		return k.do("zrevrange", k.K, start, stop, "WITHSCORES")
	}
	return k.do("zrevrange", k.K, start, stop)
}

// Zset.Rangebyscore 通过分数返回指定区间内的元素([min, max])
func (k *Zset) Rangebyscore(min, max interface{}, withScore ...bool) *reply {
	if len(withScore) > 0 && withScore[0] {
		return k.do("zrangebyscore", k.K, min, max, "WITHSCORES")
	}
	return k.do("zrangebyscore", k.K, min, max)
}

// Zset.Revrangebyscore 倒序通过分数返回指定区间内的元素([min, max])
func (k *Zset) Revrangebyscore(min, max interface{}, withScore ...bool) *reply {
	if len(withScore) > 0 && withScore[0] {
		return k.do("zrevrangebyscore", k.K, max, min, "WITHSCORES")
	}
	return k.do("zrevrangebyscore", k.K, max, min)
}

// Zset.Del 移除特定元素
func (k *Zset) Del(members ...interface{}) *reply {
	return k.do("zrem", redis.Args{}.Add(k.K).AddFlat(members)...)
}

// Zset.DelRange 移除指定区间内的元素([min_start, max_stop])
func (k *Zset) DelRange(min_start, max_stop interface{}, byScore ...bool) *reply {
	if len(byScore) > 0 && byScore[0] {
		return k.do("zremrangebyscore", k.K, min_start, max_stop)
	}
	return k.do("zremrangebyrank", k.K, min_start, max_stop)
}
