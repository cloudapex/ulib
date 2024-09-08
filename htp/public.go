package htp

import (
	"fmt"

	"github.com/cloudapex/ulib/ctl"
	"github.com/cloudapex/ulib/util"

	"github.com/gin-gonic/gin"
)

var (
	Ctl IContrler // 默认htp控制器

	units     = []IGroupRouter{}
	observers = []TObserveCallBack{}
)

// 安装控制器
func Install(conf *Config) ctl.IControler {
	c := ctl.Install(Controller(conf))
	util.Cast(Ctl == nil, func() { Ctl = c.(IContrler) }, nil)
	return c
}

// 注册路由组对象
func Register(r IGroupRouter) {
	if index := hasRouterUnit(r.Name()); index >= 0 {
		panic(fmt.Errorf("! Register IGroupRouter name:%q is already existed.", r.Name()))
	}
	units = append(units, r)
}

// 添加观察者(当调用API时回调)
func Observe(callback TObserveCallBack) {
	observers = append(observers, callback)
}

// 设置gin运行模式
func SetRunMode(mode string) {
	gin.SetMode(mode)
}
func GetRunMode() string { return gin.Mode() }

//  --------------------
func hasRouterUnit(name string) int {
	for i, unit := range units {
		if unit.Name() == name {
			return i
		}
	}
	return -1
}
