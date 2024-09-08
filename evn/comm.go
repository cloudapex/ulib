package evn

import "github.com/cloudapex/ulib/ctl"

// > 控制器接口
type IContrler interface {
	ctl.IControler

	// 监听事件(eventId重复则进行覆盖)
	Listen(event IEvent, handle TEventHandler)

	// 投递事件(sync:此时间是否需要被同步有序处理)
	Post(event IEvent, sync ...bool)

	// 投递事件并自动监听(sync:此时间是否需要被同步有序处理)
	PostDo(event IEventDo, sync ...bool)
}

// ==================== Event

// > 事件对象接口
type IEvent interface {
	EventId() TEventID
}

// > 事件(闭包)接口
type IEventDo interface {
	IEvent
	Do()
}

// 事件处理器原型
type TEventHandler func(IEvent)

// ==================== Task
type TaskHandler interface {
	OnHandleTask(param interface{}) (ret interface{}, err error)
}
type TaskHandFunc func(param interface{}) (ret interface{}, err error)

func (f TaskHandFunc) OnHandleTask(param interface{}) (ret interface{}, err error) {
	return f(param)
}
