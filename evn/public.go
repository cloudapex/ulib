package evn

import (
	"github.com/cloudapex/ulib/ctl"

	"github.com/cloudapex/ulib/util"
)

var (
	Ctl IContrler // 默认事件系统控制器

	units = map[TEventID]TEventHandler{}
)

// 安装控制器
func Install(conf *Config) ctl.IControler {
	c := ctl.Install(Controller(conf))
	util.Cast(Ctl == nil, func() { Ctl = c.(IContrler) }, nil)
	return c
}

// Register 预注册监听事件
func Register(id TEventID, handle TEventHandler) {
	units[id] = handle
}

// Listen 动态监听事件(一般情况下使用Register)
func Listen(event IEvent, handle TEventHandler) {
	Ctl.Listen(event, handle)
}

// 投递事件(orderly:此事件是否需要被有序处理)
func Post(event IEvent, orderly ...bool) {
	Ctl.Post(event, orderly...)
}

// 投递事件并自动监听(orderly:此事件是否需要被有序处理)
func PostDo(event IEventDo, orderly ...bool) {
	Ctl.PostDo(event, orderly...)
}
