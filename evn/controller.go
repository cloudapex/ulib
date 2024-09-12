package evn

import (
	"fmt"

	"github.com/cloudapex/ulib/ctl"
	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"golang.org/x/exp/rand"
)

func Controller(conf *Config) IContrler { return &controller{Conf: conf} }

// > event controller
type controller struct {
	log.ILoger
	util.RWLocker

	tasks []*Task

	handles map[TEventID]TEventHandler

	Conf *Config
}

func (this *controller) HandleName() string { return "evn" }

func (this *controller) HandleInit() {
	this.ILoger = ctl.Logger(this.HandleName())

	this.handles = map[TEventID]TEventHandler{}
	util.Cast(this.Conf == nil, func() { log.Fatal("conf = nil") }, nil)
	util.Cast(len(units) != 0, func() { this.handles = units }, nil)

	this.Conf.revise()

	// init tasks
	this.TraceD(-1, "Start add task(%d)...", this.Conf.Size)
	defer this.InfoD(-1, "Add Task(%d) done.", this.Conf.Size)

	for n := 0; n < this.Conf.Size; n++ {
		this.tasks = append(this.tasks, (&Task{}).Init(fmt.Sprintf("%s-%d", this.HandleName(), n), this.Conf.Capy))
		this.tasks[n].Handler(this)
	}
}
func (this *controller) HandleTerm() {
	for _, t := range this.tasks {
		t.Exit()
	}
}

//  ==================== Functions
// 监听事件(eventId重复则进行覆盖)
func (this *controller) Listen(event IEvent, handle TEventHandler) {
	defer this.UnLock(this.Lock())
	if _, ok := this.handles[event.EventId()]; ok {
		this.Warn("eventId:%q is already existed and reupdate it.", event.EventId())
	}
	this.handles[event.EventId()] = handle
}

// 投递事件(orderly:是否需要被有序处理)
func (this *controller) Post(event IEvent, orderly ...bool) {
	// 有序事件使用tasks[0]
	if util.DefaultVal(orderly) {
		this.tasks[0].Post(event)
		return
	}
	// 有序事件使用其他tasks
	this.tasks[rand.Intn(len(this.tasks)-1)+1].Post(event)
}

// 投递事件并自动监听(orderly:是否需要被有序处理)
func (this *controller) PostDo(event IEventDo, orderly ...bool) {
	this.Post(event, orderly...)
}

//  ==================== TaskHandle
func (this *controller) OnHandleTask(param interface{}) (ret interface{}, err error) {
	event := param.(IEvent)
	if eventDo, ok := event.(IEventDo); ok {
		eventDo.Do()
		return nil, nil
	}

	defer this.RUnLock(this.RLock())
	hander, ok := this.handles[event.EventId()]
	util.Cast(ok, func() { hander(param.(IEvent)) }, func() { this.ErrorD(-1, "EventHandle not found for event:%#v", event) })

	return nil, nil
}
