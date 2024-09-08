package htp

import (
	"github.com/cloudapex/ulib/htp/core"
	"github.com/cloudapex/ulib/htp/middleware"

	"github.com/gin-gonic/gin"
)

// ==================== JWTClaims

// CtxUserIdGet 获取user_id
func CtxUserIdGet(c *gin.Context) core.TUserID {
	var id core.TUserID
	if id, ok := c.Get(core.C_CTX_USER_ID); ok {
		return id.(core.TUserID)
	}
	return id
}

// CtxUserIdSet 设置user_id (业务层设置)
func CtxUserIdSet(c *gin.Context, uid core.TUserID) {
	c.Set(core.C_CTX_USER_ID, uid)
	// update set C_BEHAVIOR_USER_ID
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_USER_ID, uid)
}

// CtxUserNameGet 获取user_name
func CtxUserNameGet(c *gin.Context) string {
	if id, ok := c.Get(core.C_CTX_USER_NAME); ok {
		return id.(string)
	}
	return ""
}

// CtxUserNameSet 设置user_name (业务层设置)
func CtxUserNameSet(c *gin.Context, name string) {
	c.Set(core.C_CTX_USER_NAME, name)
}

// CtxRemoteIpGet  get remote_ip from Context
func CtxRemoteIpGet(c *gin.Context) string {
	if r, ok := c.Get(core.C_CTX_REMOTE_IP); ok {
		return r.(string)
	}
	return ""
}

// CtxRemoteIpSet  set remote_ip to Context (业务层设置)
func CtxRemoteIpSet(c *gin.Context, ip string) {
	c.Set(core.C_CTX_REMOTE_IP, ip)

	// update set middleware.Behavior
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_CLIENT_IP, ip)
}

// CtxRequestGet  get request from Context
func CtxRequestGet(c *gin.Context) any {
	if r, ok := c.Get(core.C_CTX_REQUEST); ok {
		return r
	}
	return nil
}

// CtxRequestSet  set request to Context (系统内部或中间件设置)
func CtxRequestSet(c *gin.Context, req any) {
	c.Set(core.C_CTX_REQUEST, req)

	// update set middleware.Behavior
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_REQUEST, req)
}

// CtxResponseGet  get response from Context
func CtxResponseGet(c *gin.Context) *Response {
	if r, ok := c.Get(core.C_CTX_RESPONSE); ok {
		return r.(*Response)
	}
	return nil
}

// CtxGetResponse  set response to Context (系统内部或中间件设置)
func CtxResponseSet(c *gin.Context, rsp *Response) {
	c.Set(core.C_CTX_RESPONSE, rsp)

	// update set middleware.Behavior
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_CODE, rsp.Code)
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_RESPONSE, rsp)
	middleware.BehaviorSet(c, middleware.C_BEHAVIOR_RESP_DATA, rsp.Data)
}
