package evn

import (
	"fmt"
	"time"

	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"
)

// Task
type Task struct {
	name   string
	exit   chan int
	queue  chan *task
	handle TaskHandler
}

func (this *Task) Init(name string, capy int) *Task {
	this.name = name
	this.exit, this.queue = make(chan int), make(chan *task, capy)
	return this
}
func (this *Task) Handler(handler TaskHandler) {
	util.Cast(this.handle == nil, func() { this.handle = handler; go this.loop() }, nil)
}
func (this *Task) HandleFunc(handFun TaskHandFunc) {
	this.Handler(handFun)
}
func (this *Task) Post(param interface{}, donefun_ ...func(result interface{}, err error)) {
	var donefun func(result interface{}, err error)
	if len(donefun_) > 0 {
		donefun = donefun_[0]
	}
	util.Cast(this.handle != nil, func() { this.queue <- &task{param, donefun} }, nil)
}
func (this *Task) Exit() {
	if this.handle == nil {
		return
	}
	select {
	case this.exit <- 1:
		return
	case <-time.After(C_TASK_EXIT_TIME_OUT):
	}
	log.ErrorD(-1, "Task[%q] Exit Timeout", this.name)
}
func (this *Task) loop() {
	defer func() {
		if util.Catch(fmt.Sprintf("Task[%q] panic but it will resume", this.name), recover()) {
			go this.loop()
		}
	}()

	for {
		select {
		case <-this.exit:
			return
		case t := <-this.queue:
			ret, err := this.handle.OnHandleTask(t.param)
			if t.donefun == nil {
				if err != nil {
					log.ErrorD(-1, "Task[%q] handle task(%#v) err:%v", this.name, t.param, err)
				}
				continue
			}
			t.donefun(ret, err)
		}
	}
}
