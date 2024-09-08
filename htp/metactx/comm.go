package metactx

import (
	"github.com/cloudapex/ulib/htp/core"

	"github.com/gin-gonic/gin"
)

// > MetaCtx 接口
type IContext interface {
	Ctx() *gin.Context

	UserID() core.TUserID
	UserName() string
	ClientIP() string
	Language() string

	Head(headKey string) string

	Set(obj interface{}, flag ...interface{})
	Get(nilStruct interface{}, flag ...interface{}) interface{}
}
