package rdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-05-13
Description: 回应结构
*******************************************************************************/
import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudapex/ulib/log"

	"github.com/gomodule/redigo/redis"
	"github.com/vmihailenco/msgpack"
	"google.golang.org/protobuf/proto"
)

// ==================== Coding

func init() {
	RegistCoder(ECod_Json, Coder{json.Marshal, json.Unmarshal})
	RegistCoder(ECod_Msgpack, Coder{
		func(v interface{}) ([]byte, error) { return msgpack.Marshal(v) },
		func(data []byte, v interface{}) error { return msgpack.Unmarshal(data, v) },
	})
	RegistCoder(ECod_ProtoBuff, Coder{
		func(v interface{}) ([]byte, error) { return proto.Marshal(v.(proto.Message)) },
		func(data []byte, v interface{}) error { return proto.Unmarshal(data, v.(proto.Message)) },
	})
}

func Encode(encoding ECoding, v interface{}) (interface{}, error) {
	if encoding == ECod_None {
		return v, nil
	}
	coder, ok := mapCoders[encoding]
	if !ok {
		return nil, ErrInvalidCoding
	}
	return coder.Encoder(v)
}

func Decode(encoding ECoding, b []byte, v interface{}) error {
	if encoding == ECod_None {
		return nil
	}
	coder, ok := mapCoders[encoding]
	if !ok {
		return ErrInvalidCoding
	}
	return coder.Decoder(b, v)
}

// ==================== Reply

type Resp = reply

func Reply(rep interface{}, err error, coding ECoding, fullCmd string) *reply {
	if err != nil && err != redis.ErrNil {
		log.Error("Reply command: %v, err = %v", fullCmd, err)
	}
	return &reply{rep: rep, err: err, coding: coding}
}

type reply struct {
	rep    interface{}
	err    error
	coding ECoding
	strict bool
	ext    interface{} // 额外辅助数据
}

func (r *reply) Result() (interface{}, error) {
	return r.rep, r.err
}

func (r *reply) Error() error {
	return r.err
}

// Strict 严格错误(默认 redis.ErrNil 返回nil, Strict()之后返回error)
func (r *reply) Strict() *reply {
	r.strict = true
	return r
}

func (r *reply) Int() (int, error) {
	val, err := redis.Int(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return 0, nil
	}
	return val, err
}

func (r *reply) Int64() (int64, error) {
	val, err := redis.Int64(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return 0, nil
	}
	return val, err
}

func (r *reply) Uint64() (uint64, error) {
	val, err := redis.Uint64(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return 0, nil
	}
	return val, err
}

func (r *reply) Float64() (float64, error) {
	val, err := redis.Float64(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return 0, nil
	}
	return val, err
}

func (r *reply) String() (string, error) {
	val, err := redis.String(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return "", nil
	}
	return val, err
}

func (r *reply) Bytes() ([]byte, error) {
	val, err := redis.Bytes(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) Unmarshal(v interface{}) error {
	b, err := r.Bytes()
	if !r.strict && err == redis.ErrNil {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return Decode(r.coding, b, v)
}
func (r *reply) UnmarshalJSArray(v interface{}, replaceNilTo ...string) error {
	arr, err := r.Strings()
	if !r.strict && err == redis.ErrNil {
		return nil
	}
	if err != nil {
		return err
	}

	nilTo := "{}"
	if len(replaceNilTo) > 0 {
		nilTo = replaceNilTo[0]
	}
	for n := range arr {
		if arr[n] == "" || strings.Contains(arr[n], "nil") {
			arr[n] = nilTo
		}
	}

	return Decode(ECod_Json, []byte(fmt.Sprintf("[%s]", strings.Join(arr, ","))), v)
}
func (r *reply) Bool() (bool, error) {
	val, err := redis.Bool(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return false, nil
	}
	return val, err
}

func (r *reply) Float64s() ([]float64, error) {
	val, err := redis.Float64s(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) Strings() ([]string, error) {
	val, err := redis.Strings(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) ByteSlices() ([][]byte, error) {
	val, err := redis.ByteSlices(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) Int64s() ([]int64, error) {
	val, err := redis.Int64s(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) Ints() ([]int, error) {
	val, err := redis.Ints(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) HashMap() (map[interface{}]string, error) {
	vals, err := redis.Strings(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return map[interface{}]string{}, nil
	}
	if err != nil {
		return nil, err
	}

	fileds := r.ext.([]interface{})
	if len(vals) != len(fileds) {
		return nil, ErrHashGetMapResultMismatch
	}
	ret := map[interface{}]string{}
	for n, v := range vals {
		ret[fileds[n]] = v
	}
	return ret, nil
}

func (r *reply) StringMap() (map[string]string, error) {
	val, err := redis.StringMap(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return map[string]string{}, nil
	}
	return val, err
}

func (r *reply) IntMap() (map[string]int, error) {
	val, err := redis.IntMap(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return map[string]int{}, nil
	}
	return val, err
}

func (r *reply) Int64Map() (map[string]int64, error) {
	val, err := redis.Int64Map(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return map[string]int64{}, nil
	}
	return val, err
}

func (r *reply) Positions() ([]*[2]float64, error) {
	val, err := redis.Positions(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) Values() ([]interface{}, error) {
	val, err := redis.Values(r.rep, r.err)
	if !r.strict && err == redis.ErrNil {
		return nil, nil
	}
	return val, err
}

func (r *reply) ScanStruct(pObj interface{}) error {
	v, err := r.Values()
	if !r.strict && err == redis.ErrNil {
		return nil
	} else if err != nil {
		return err
	}
	return redis.ScanStruct(v, pObj)
}
