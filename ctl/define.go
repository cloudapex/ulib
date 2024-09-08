package ctl

import "time"

const (
	C_TIMER_TICK_INTERVAL = 1 * time.Minute // 定时器默认tick精度
)

// > IControl接口
type IControler interface {

	// 控制器名称
	HandleName() string

	// 控制器准备
	HandleInit()

	// 控制器销毁
	HandleTerm()
}

// > appInfo
type appInfo struct {
	Name    string
	Version string // x.y.z
}

// ==================== Timer

// 状态存储接口
type ITimerRestorer interface {

	// 恢复状态
	Load() map[string]int64

	// 保存状态
	Save(key string, valAt int64)
}

// 定时器接口
type ITimer interface {
	// 初始化
	Init(tickerInterval ...time.Duration) *timer

	// 启动
	Start(restorer ...ITimerRestorer) *timer

	// 关闭
	Close()

	// 设置并恢复状态(需要在Handler添加之后调用)
	Restore(restorer ITimerRestorer)

	// 添加定时器[当fireFun返回false时自动删除此TimerHandler]
	TimerHandler(d time.Duration, fireFun TTimerHandFunc, opt ...*TimerOpt)

	// 添加每天一次的定时器[同上](after0:每天超过零点多少时间)
	DailyHandler(after0 time.Duration, fireFun TTimerHandFunc, opt ...*TimerOpt)

	// 移除定时器
	DelHandler(handle TTimerHandFunc)
	DelHandlerByName(name string)
}
type TTimerHandFunc func(now time.Time) (keep bool)
