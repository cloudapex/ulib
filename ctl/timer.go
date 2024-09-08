package ctl

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cloudapex/ulib/util"
)

func Timer() ITimer { return &timer{} }

// timer
type timer struct {
	mutex    util.RWLocker
	cronjobs map[string]*cronjob

	interval time.Duration // 定时器tick时间间隔

	sgExit   chan int
	wgExit   sync.WaitGroup
	restorer ITimerRestorer // 外部状态存储器
}

func (this *timer) Init(tickerInterval ...time.Duration) *timer {
	this.cronjobs = make(map[string]*cronjob)

	this.interval = util.DefaultVal(tickerInterval)
	util.Cast(this.interval == 0, func() { this.interval = C_TIMER_TICK_INTERVAL }, nil)

	return this
}
func (this *timer) Start(restorer ...ITimerRestorer) *timer {
	this.sgExit = make(chan int)

	util.Cast(len(restorer) > 0 && restorer[0] != nil, func() { this.Restore(restorer[0]) }, nil)

	return func() *timer { go this.loop(); return this }()
}
func (this *timer) Close() {
	util.Cast(this.sgExit != nil, func() { this.sgExit <- 0 }, nil)

	this.wgExit.Wait()
}

// 设置并恢复状态(需要在Handler添加之后调用)
func (this *timer) Restore(restorer ITimerRestorer) {
	this.restorer = restorer

	for name, last := range restorer.Load() {
		for name_, it := range this.cronjobs {
			if name_ == name {
				it.last = time.Unix(last, 0)
				break
			}
		}
	}
}
func (this *timer) TimerHandler(interval time.Duration, handle TTimerHandFunc, opt ...*TimerOpt) {
	this.addHandler(interval, handle, false, opt...)
}
func (this *timer) DailyHandler(absolute time.Duration, handle TTimerHandFunc, opt ...*TimerOpt) {
	this.addHandler(absolute, handle, true, opt...)
}
func (this *timer) DelHandler(handle TTimerHandFunc) {
	defer this.mutex.UnLock(this.mutex.Lock())
	delete(this.cronjobs, util.FuncFullNameRef(reflect.ValueOf(handle)))
}
func (this *timer) DelHandlerByName(name string) {
	defer this.mutex.UnLock(this.mutex.Lock())
	delete(this.cronjobs, name)
}

// --------------------

func (this *timer) tick() {
	defer this.mutex.RUnLock(this.mutex.RLock())
	if len(this.cronjobs) == 0 {
		return
	}

	curTime := time.Now()
	for name, t := range this.cronjobs {
		lstTime := t.last

		if t.daily {
			if curTime.YearDay() == t.last.YearDay() {
				continue
			}
			lstTime = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, curTime.Location())
		}

		if curTime.Sub(lstTime) <= t.interval {
			continue
		}

		t.last = curTime
		if t.store && this.restorer != nil {
			this.restorer.Save(t.name, t.last.Unix())
		}

		_name, _fun := name, t.fun
		util.Goroutine(fmt.Sprintf("cronjob[%q]", _name), func() {
			if ret := _fun.Call([]reflect.Value{reflect.ValueOf(curTime)}); !ret[0].Bool() {
				this.DelHandlerByName(_name)
			}
		}, &this.wgExit)
	}
}
func (this *timer) loop() {
	t := time.NewTicker(this.interval)

	defer func() {
		if util.Catch("Timer.loop() panic and it will resume", recover()) {
			go this.loop()
		}
	}()

	for {
		select {
		case <-t.C:
			this.tick()
		case <-this.sgExit:
			return
		}
	}
}
func (this *timer) addHandler(interval time.Duration, handle TTimerHandFunc, daily bool, opt ...*TimerOpt) {
	var _opt TimerOpt
	if len(opt) > 0 && opt[0] != nil {
		_opt = *opt[0]
	}

	last := time.Now()
	if _opt.Right {
		if daily { // 当天时间只要满足则执行,否则次日才会开始执行
			last = time.Time{}
		} else if !handle(time.Now()) { // 立即执行
			return
		}
	}

	t := &cronjob{_opt.Name, last, daily, _opt.Store, interval, reflect.ValueOf(handle)}
	if t.name == "" {
		t.name = util.FuncFullNameRef(t.fun)
	}

	defer this.mutex.UnLock(this.mutex.Lock())
	this.cronjobs[t.name] = t
}

// ==================== cronjob
type cronjob struct {
	name         string        // 定时任务的名称
	last         time.Time     // 上次执行的时间
	daily, store bool          // 是否是日常类型,是否需要保存状态
	interval     time.Duration // 调用间隔时间
	fun          reflect.Value // 调用函数
}

// ==================== TimerOpt(可选参数)
type TimerOpt struct {
	Name  string // 自定义名称(默认为函数名util.FuncFullNameRef)
	Right bool   // 是否立刻执行(daily类:当天时间满足的话则执行)
	Store bool   // 是否需要保存状态
}
