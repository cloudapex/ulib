package evn

import (
	"time"

	"github.com/duke-git/lancet/v2/mathutil"
)

const C_TASK_EXIT_TIME_OUT = 2 * time.Second // task  退出超时

type Config struct {
	Size int // 处理事件的任务数量
	Capy int // 每个任务的通道能力
} //
func (c *Config) revise() {
	c.Size = mathutil.Max(c.Size, 5)
	c.Capy = mathutil.Max(c.Capy, 100)
}

// 事件Id类型
type TEventID = string

// 事件基本结构(供业务简单使用)
type Event struct {
	Id TEventID
}                                  //
func (e *Event) EventId() TEventID { return e.Id }

// 事件(闭包)结构(供业务继承)
func EventDoFun(do func()) EventDo { return EventDo{Fun: do} }

type EventDo struct {
	Id  TEventID
	Fun func()
}                                    //
func (e *EventDo) Do()               { e.Fun() }
func (e *EventDo) EventId() TEventID { return e.Id }

// Task结构
type task struct {
	param   interface{}
	donefun func(result interface{}, err error)
}
