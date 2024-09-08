package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudapex/ulib/htp/core"
	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/gin-gonic/gin"
)

// 管理behavior的context中的key
const (
	C_CTX_KEY_BEHAVIOR   = "behavior"
	C_BEHAVIOR_COST_WARN = 500 * time.Millisecond // api耗时超过阀值则log等级提升为warn
)

// 管理behavior的字段数据类型
type TBehaviorField = string // 包里内置的几个字段(业务层可扩展)
const (
	c_behavior_ignore TBehaviorField = "_ignore" // 用来标记本次请求忽略行为分析

	C_BEHAVIOR_CLIENT_IP  TBehaviorField = "client_ip"  // client ip
	C_BEHAVIOR_USER_ID    TBehaviorField = "uid"        // user id (string)(业务层)
	C_BEHAVIOR_REQUEST_ID TBehaviorField = "request_id" // request id (需请求头中含有C_REQ_HEAD_REQUEST_ID)
	C_BEHAVIOR_API        TBehaviorField = "api"        // api
	C_BEHAVIOR_METHOD     TBehaviorField = "method"     // Http Method
	C_BEHAVIOR_COST       TBehaviorField = "cost"       // 耗时(ms)
	C_BEHAVIOR_COST_TRACE TBehaviorField = "cost_trace" // 耗时追踪(ms)(业务层)
	C_BEHAVIOR_RETRY_AT   TBehaviorField = "retry_at"   // 重试时间((需请求头中含有C_REQ_HEAD_RETRY_AT))
	C_BEHAVIOR_REQ_HEAD   TBehaviorField = "head"       // 请求头
	C_BEHAVIOR_REQUEST    TBehaviorField = "request"    // 请求数据
	C_BEHAVIOR_STATUS     TBehaviorField = "status"     // 回应http状态码(int)
	C_BEHAVIOR_CODE       TBehaviorField = "code"       // 回应业务码(int)
	C_BEHAVIOR_RSPSIZE    TBehaviorField = "resp_size"  // 回应字节大小(KB)

	C_BEHAVIOR_RESPONSE  TBehaviorField = "response"      // 回应数据(不单独占用logger.field,使用message作为输出)
	C_BEHAVIOR_RESP_DATA TBehaviorField = "response_data" // 回应数据中的Data字段(不单独占用logger.field,仅逻辑用途)
)

// behavior set field's value
func BehaviorSet(c *gin.Context, field TBehaviorField, val interface{}) {
	if v, ok := c.Get(C_CTX_KEY_BEHAVIOR); ok {
		v.(map[TBehaviorField]interface{})[field] = val
	}
}

// behavior get field's value
func BehaviorGet(c *gin.Context, field TBehaviorField) interface{} {
	if v, ok := c.Get(C_CTX_KEY_BEHAVIOR); ok {
		if val, exist := v.(map[TBehaviorField]interface{})[field]; exist {
			return val
		}
	}
	return nil
}

// behavior will be ignore
func BehaviorIgnore(c *gin.Context) {
	BehaviorSet(c, c_behavior_ignore, true)
}

// behavior set cost_trace to add trace sub module of cost duration
func HeavierCostTrace(c *gin.Context, traceModule string, d time.Duration) {
	if v, ok := c.Get(C_CTX_KEY_BEHAVIOR); ok {
		values := v.(map[TBehaviorField]interface{})
		if _, existed := values[C_BEHAVIOR_COST_TRACE]; !existed {
			values[C_BEHAVIOR_COST_TRACE] = []string{}
		}
		values[C_BEHAVIOR_COST_TRACE] = append(values[C_BEHAVIOR_COST_TRACE].([]string), fmt.Sprintf("%s-%d", traceModule, d.Milliseconds()))
	}
}

// > 中间件[Behavior]
func Behavior(debugModel bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// init behavior context kvs
		c.Set(C_CTX_KEY_BEHAVIOR, map[TBehaviorField]interface{}{})

		// fill behavior base field
		BehaviorSet(c, C_BEHAVIOR_CLIENT_IP, c.ClientIP())
		BehaviorSet(c, C_BEHAVIOR_REQUEST_ID, c.Request.Header.Get(core.C_HTTP_HEAD_REQ_ID))
		BehaviorSet(c, C_BEHAVIOR_API, c.Request.URL.Path)
		BehaviorSet(c, C_BEHAVIOR_METHOD, c.Request.Method)
		BehaviorSet(c, C_BEHAVIOR_RETRY_AT, c.Request.Header.Get(core.C_HTTP_HEAD_RETRY_AT))
		BehaviorSet(c, C_BEHAVIOR_REQ_HEAD, requestHead(c))
		BehaviorSet(c, C_BEHAVIOR_REQUEST, requestData(c)) // C_BEHAVIOR_REQUEST 可能会被业务层给重置,下面会再取一次
		// fill default value
		BehaviorSet(c, C_BEHAVIOR_CODE, -1)
		BehaviorSet(c, C_BEHAVIOR_RSPSIZE, -1)

		defer func(begin time.Time) {
			val, _ := c.Get(C_CTX_KEY_BEHAVIOR)
			behaviors := val.(map[string]interface{})

			// ignore
			if v, ok := behaviors[c_behavior_ignore]; ok && v.(bool) {
				return
			}

			// C_BEHAVIOR_REQUEST
			req := behaviors[C_BEHAVIOR_REQUEST]
			if _, ok := req.(string); !ok { // reset for req struct
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetEscapeHTML(false)
				enc.Encode(req)
				BehaviorSet(c, C_BEHAVIOR_REQUEST, fmt.Sprintf("%s:%s", util.StructName(req), buf.String()))
			}

			// C_BEHAVIOR_COST
			cost := time.Since(begin)
			BehaviorSet(c, C_BEHAVIOR_COST, cost.Milliseconds())

			// C_BEHAVIOR_COST_TRACE(可选)
			if v, ok := behaviors[C_BEHAVIOR_COST_TRACE]; ok {
				BehaviorSet(c, C_BEHAVIOR_COST_TRACE, fmt.Sprintf("[%s]", strings.Join(v.([]string), ", ")))
			}

			// C_BEHAVIOR_STATUS
			status := c.Writer.Status()
			BehaviorSet(c, C_BEHAVIOR_STATUS, status)

			// C_BEHAVIOR_RSPSIZE
			BehaviorSet(c, C_BEHAVIOR_RSPSIZE, c.Writer.Size())

			// C_BEHAVIOR_RESPONSE
			rspdata := "nil"
			if v, ok := behaviors[C_BEHAVIOR_RESPONSE]; ok {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetEscapeHTML(false)
				enc.Encode(v)
				rspdata = fmt.Sprintf("%s", buf.String())
			}

			// logLevel
			isErr := false
			isWan := false
			if status != http.StatusOK {
				isErr = true
			} else if code, _ := behaviors[C_BEHAVIOR_CODE].(int); code != 0 {
				if behaviors[C_BEHAVIOR_RESP_DATA] != nil {
					isWan = true
				} else {
					isErr = true
				}
			}

			// logger
			delete(behaviors, C_BEHAVIOR_RESPONSE)
			delete(behaviors, C_BEHAVIOR_RESP_DATA)
			l := log.Fields(behaviors)
			if isErr {
				l.ErrorD(-1, "[API] %s response:%s", c.Request.URL.Path, rspdata)
			} else if isWan || cost >= C_BEHAVIOR_COST_WARN {
				l.WarnD(-1, "[API] %s response:%s", c.Request.URL.Path, rspdata)
			} else if debugModel {
				l.DebugD(-1, "[API] %s response:%s", c.Request.URL.Path, rspdata)
			}

		}(time.Now())

		c.Next()
	}
}

//------------------------------------------------------------------------------

func requestHead(c *gin.Context) string {
	head, _ := json.Marshal(c.Request.Header)
	return string(head)
}

func requestData(c *gin.Context) string {
	query, _ := url.QueryUnescape(c.Request.URL.RawQuery)
	body, _ := c.GetRawData()

	req := fmt.Sprintf("%s >|< %s", query, body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return req
}
