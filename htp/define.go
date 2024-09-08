package htp

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/cloudapex/ulib/log"

	"github.com/gin-gonic/gin"
)

// > 配置项
type Config struct {
	RunMode      string    `json:"runMode"` // debug release
	ListenAddr   string    `json:"listenAddr"`
	WriteTimeout int       `json:"writeTimeout"` // second
	ReadTimeout  int       `json:"readTimeout"`  // second
	ListnTls     ListenTLS `json:"listnTls"`
}
type ListenTLS struct {
	Enable  bool   `json:"enable"`
	CrtFile string `json:"crtFile"`
	KeyFile string `json:"keyFile"`
}

// > 路由服务渲染模式
type ESRenderMode int //
const (
	ESRender_None ESRenderMode = iota // 以xml数据格式返回
	ESRender_Xml                      // 以xml数据格式返回
	ESRender_Yaml                     // 以yaml数据格式返回
	ESRender_Json                     // 以JSON数据格式返回
	ESRender_Pbuf                     // 以Protobuf数据格式返回
) // Inherit from fmt.Stringer interface
func (e ESRenderMode) String() string {
	switch e {
	case ESRender_Xml:
		return "ESRender_Xml"
	case ESRender_Yaml:
		return "ESRender_Yaml"
	case ESRender_Json:
		return "ESRender_Json"
	case ESRender_Pbuf:
		return "ESRender_Pbuf"
	}
	return "UnKnow"
}

// > GroupRouter
type GroupRouter struct {
	*gin.RouterGroup
}

// 返回自身的RouterGroup
func (gr *GroupRouter) Base() *gin.RouterGroup { return gr.RouterGroup }

// 注册API接口服务(GET|POST)
func (gr *GroupRouter) API(group *gin.RouterGroup, relativePath string, service IService, otherHandlers ...gin.HandlerFunc) {
	var handlers []gin.HandlerFunc
	serviceT := reflect.TypeOf(service).Elem()

	handlers = append([]gin.HandlerFunc{func(c *gin.Context) { Service(c, reflect.New(serviceT).Interface().(IService)) }},
		otherHandlers...)

	group.GET(relativePath, handlers...)
	group.POST(relativePath, handlers...)

	// observers callback
	path := path.Join(group.BasePath(), relativePath)
	for _, callback := range observers {
		callback(path, service)
	}
}

// -------- GinLogger
type GinLogger struct{} //
func (this GinLogger) Write(p []byte) (n int, err error) {
	info := string(p)
	if bytes.HasPrefix(p, []byte(fmt.Sprintf("[GIN] %d/", time.Now().Year()))) {
		info = "[GIN]" + info[28:]
	}
	log.DebugD(-1, strings.TrimSpace(info))
	return len(p), nil
}

// -------- GinRecover
type GinRecover struct{} //
func (this GinRecover) Write(p []byte) (n int, err error) {
	log.ErrorD(-1, string(p))
	return len(p), nil
}
