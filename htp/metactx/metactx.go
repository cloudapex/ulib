package metactx

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/cloudapex/ulib/htp/core"

	"github.com/gin-gonic/gin"
)

// 构造函数1: 请求链场景
func WithCtx(c *gin.Context) IContext {
	return &metaCtx{
		ctx:     c,
		heads:   map[string]interface{}{},
		objects: map[string]interface{}{},
	}
}

// 构造函数2: 非请求链场景
func WithUser(uid core.TUserID) IContext {
	return &metaCtx{
		ctx:     nil,
		uid:     uid,
		heads:   map[string]interface{}{},
		objects: map[string]interface{}{},
	}
}

// 元数据上下文
type metaCtx struct {
	ctx *gin.Context
	// uid
	uid core.TUserID
	// Heads
	heads map[string]interface{}
	// Context with objects
	objects map[string]interface{}
}

// 取得gin.Context
func (m *metaCtx) Ctx() *gin.Context { return m.ctx }

// 取得client的用户ID
func (m *metaCtx) UserID() core.TUserID {
	var id core.TUserID
	if !core.IsZeroUID(m.uid) {
		return m.uid
	}
	if id, ok := m.Ctx().Get(core.C_CTX_USER_ID); ok {
		return id.(core.TUserID)
	}
	return id
}

// 取得client的用户昵称
func (m *metaCtx) UserName() string {
	if id, ok := m.Ctx().Get(core.C_CTX_USER_NAME); ok {
		return id.(string)
	}
	return ""
}

// 取得client的远端IP
func (m *metaCtx) ClientIP() string {
	if r, ok := m.Ctx().Get(core.C_CTX_REMOTE_IP); ok {
		return r.(string)
	}
	return m.Ctx().ClientIP()
}

// 取得client所用语言
func (m *metaCtx) Language() string {
	return metaHeadValue(m.ctx, m.heads, core.C_HTTP_HEAD_LANGUAGE, conv_to_string)
}

// 取得head-value
func (m *metaCtx) Head(headKey string) string {
	return metaHeadValue(m.ctx, m.heads, headKey, conv_to_string)
}

// 管理缓存对象
func (m *metaCtx) Set(obj interface{}, flag ...interface{}) {
	m.objects[fmt.Sprintf("%s-%v", reflect.TypeOf(obj).String(), flag)] = obj
}
func (m *metaCtx) Get(nilStruct interface{}, flag ...interface{}) interface{} {
	return m.objects[fmt.Sprintf("%s-%v", reflect.TypeOf(nilStruct).String(), flag)]
}

// --------------- internal

func metaHeadValue[T any](c *gin.Context, heads map[string]interface{}, headKey string, converter func(src string) T) T {
	v, exist := heads[headKey]
	if exist {
		val, _ := v.(T)
		return val
	}
	var val T
	value := c.Request.Header.Get(headKey)
	if len(value) > 0 {
		val = converter(value)
	}
	heads[headKey] = val
	return val
}

func conv_to_int(strNum string) int {
	n, _ := strconv.Atoi(strNum)
	return n
}
func conv_to_int64(strNum string) int64 {
	n, _ := strconv.ParseInt(strNum, 10, 64)
	return n
}
func conv_to_string(src string) string { return src }
