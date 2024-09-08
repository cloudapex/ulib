package log

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	C_LOG_MODE        = ELM_Std          // 默认日志输出模式
	C_LOG_LEVEL       = ELL_Debug        // 默认日志过滤等级
	C_LOG_FILE_SUFFIX = "log"            // 默认日志文件后缀
	C_LOG_ROTATE_NUM  = 3                // 默认日志文件轮换数量
	C_LOG_ROTATE_SIZE = 20 * 1024 * 1024 // 默认日志文件轮换size
	C_LOG_CSIZE       = 2048             // 默认日志消息ChanSize

	C_TH_CHAN_OVERLOAD       = "Threshold:%s"    // 消息积压阀值名称
	C_TH_CHAN_OVERLOAD_VALUE = C_LOG_CSIZE * 0.8 // 消息积压阀值(过大时告警)
)

var (
	LOG_MSG_LV_PREFIXS = [ELL_Max]string{"[TRC]", "[DBG]", "[INF]", "[WRN]", "[ERR]", "[FAL]"} // fail
	LOG_MSG_COLORS     = [ELL_Max]int{97, 94, 92, 93, 91, 95}                                  // colors
)

// ==================== 类型定义

// 日志等级
type ELogLevel int //
const (
	ELL_Trace ELogLevel = iota
	ELL_Debug
	ELL_Infos
	ELL_Warns
	ELL_Error
	ELL_Fatal
	ELL_Max // 6
) //
func (e ELogLevel) String() string {
	if e >= ELL_Trace && e < ELL_Max {
		return LOG_MSG_LV_PREFIXS[e]
	}
	return fmt.Sprintf("ELL_Unkonw(%d)", e)
}

// 日志运行状态
type ELoggerStatus int //
const (
	ELS_Initing ELoggerStatus = iota
	ELS_Running
	ELS_Exiting
	ELS_Stopped
	ELS_Max
) //
func (e ELoggerStatus) String() string {
	switch e {
	case ELS_Initing:
		return "Initing"
	case ELS_Running:
		return "Running"
	case ELS_Exiting:
		return "Exiting"
	case ELS_Stopped:
		return "Stopped"
	}
	return fmt.Sprintf("ELS_Unkonw(%d)", e)
}

// 日志输出模式
type ELogMode int //
const (
	ELM_Std ELogMode = 1 << iota
	ELM_File
	ELM_Max
) //
func (e ELogMode) String() string {
	var str = []string{}
	if e&ELM_Std != 0 {
		str = append(str, "Std")
	}
	if e&ELM_File != 0 {
		str = append(str, "File")
	}
	return strings.Join(str, "+")
}

// ==================== 接口定义

// ILoger interface
type ILoger interface {

	// 添加扩展字段
	Field(field string, val interface{}) ILoger
	// 克隆ILoger对象
	Clone() ILoger

	// TRACE
	Trace(format string, v ...interface{})
	Tracev(v ...interface{})
	TraceD(depth int, format string, v ...interface{})
	TraceDv(depth int, v ...interface{})

	// DEBUG
	Debug(format string, v ...interface{})
	Debugv(v ...interface{})
	DebugD(depth int, format string, v ...interface{})
	DebugDv(depth int, v ...interface{})

	// INFO
	Info(format string, v ...interface{})
	Infov(v ...interface{})
	InfoD(depth int, format string, v ...interface{})
	InfoDv(depth int, v ...interface{})

	// WARN
	Warn(format string, v ...interface{})
	Warnv(v ...interface{})
	WarnD(depth int, format string, v ...interface{})
	WarnDv(depth int, v ...interface{})

	// ERROR
	Error(format string, v ...interface{})
	Errorv(v ...interface{})
	ErrorD(depth int, format string, v ...interface{})
	ErrorDv(depth int, v ...interface{})
	// FATAL
	Fatal(format string, v ...interface{})
	Fatalv(v ...interface{})
	FatalD(depth int, format string, v ...interface{})
	FatalDv(depth int, v ...interface{})
}

// ==================== 结构定义

// 日志单元
type LogUnit struct {
	Lv     ELogLevel
	Str    string
	At     time.Time
	Fields map[string]interface{}
}

// 日志配置
type Config struct {
	Level      ELogLevel `json:"lv"`         // 日志等级[ELL_Debug]
	OutMode    ELogMode  `json:"mode"`       // 日志输出模式
	DirName    string    `json:"dir"`        // 输出目录[默认在程序所在目录]
	FileName   string    `json:"fileName"`   // 日志文件主名[程序本身名]
	FileSuffix string    `json:"fileSuffix"` // 日志文件后缀[log]
	RotateMax  int       `json:"rotateMax"`  // 日志文件轮换数量[3]
	RotateSize int       `json:"rotateSize"` // 日志文件轮换大小[20m]
}

// 灰日志配置
type GraylogConf struct {
	Address       string `json:"addr"`      // graylog地址(ip:port)
	GelfIntercept bool   `json:"intercept"` // graylog是否拦截标准输出
	Service       string `json:"service"`   // app.server.env
	WithFull      bool   `json:"withFull"`  // false
}

// ==================== Threshold (阀值报警)
var mapThresholds = make(map[string]*threshold)

func Threshold(name string) *threshold {
	if m, ok := mapThresholds[name]; ok {
		return m
	}
	return nil
}
func RegThreshold(name string, referValue int64, durationRate time.Duration, fmtContent string) {
	mapThresholds[name] = &threshold{
		name:       name,
		refer:      referValue,
		interval:   durationRate,
		lastAlarm:  time.Now().Add(durationRate * -1),
		fmtContent: fmtContent,
	}
}

type threshold struct {
	name       string // 模块名
	mux        sync.RWMutex
	refer      int64         // 临界值
	incrVal    int64         // 自增值
	interval   time.Duration // 检测告警频率
	lastAlarm  time.Time     // 上次触发时间
	fmtContent string        // 告警内容
}

// 断言阀值
func (t *threshold) Assert(value int64, v ...interface{}) {
	t.mux.RLock()
	if value < t.refer || time.Since(t.lastAlarm) < t.interval {
		t.mux.RUnlock()
		return
	}
	t.mux.RUnlock()

	t.mux.Lock()
	defer t.mux.Unlock()
	t.lastAlarm = time.Now()
	WarnD(-1, fmt.Sprintf("Threshold_Assert[%s][%d/%d] ", t.name, value, t.refer)+fmt.Sprintf(t.fmtContent, v...))
}

// 增加自增值(默认+1)
func (t *threshold) IncrVal(incr int64, v ...interface{}) {
	t.mux.Lock()
	defer t.mux.Unlock()

	inc := int64(1)
	if incr > 0 {
		inc = incr
	}
	t.incrVal += inc
	if t.incrVal < t.refer || time.Since(t.lastAlarm) < t.interval {

		return
	}

	WarnD(-1, fmt.Sprintf("Threshold_Increment[%s][%d/%d] ", t.name, t.incrVal, t.refer)+fmt.Sprintf(t.fmtContent, v...))
	t.incrVal, t.lastAlarm = 0, time.Now()
}
