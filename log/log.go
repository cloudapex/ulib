package log

/******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2014-10-18
Description: 日志系统
******************************************************************************/
import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func New(conf *Config) *logger { return (&logger{}).init(conf) }

type logger struct {
	status ELoggerStatus

	level            ELogLevel
	outMode          ELogMode
	dirName          string
	fileName         string
	fileSuffix       string
	rotateMax        int
	rotateSize       int
	levelPrefixNames [ELL_Max]string
	filters          []func(msg *LogUnit) bool

	screenLogger    *log.Logger
	fileSystmHandle *os.File
	fileSystmLogger *log.Logger
	fileLogicHandle *os.File
	fileLogiclogger *log.Logger

	chanMsgs chan *LogUnit
	chanExit chan int
	wgExit   sync.WaitGroup

	lastUpdateTime time.Time
}

func (this *logger) init(conf *Config) *logger {

	this.status = ELS_Initing
	this.screenLogger = log.New(os.Stderr, "", log.Lmsgprefix)

	// default
	this.outMode, this.level = C_LOG_MODE, C_LOG_LEVEL
	this.dirName = "."
	this.fileName = strings.Split(filepath.Base(os.Args[0]), ".")[0]
	this.fileSuffix = C_LOG_FILE_SUFFIX
	this.rotateMax, this.rotateSize = C_LOG_ROTATE_NUM, C_LOG_ROTATE_SIZE
	this.levelPrefixNames = LOG_MSG_LV_PREFIXS
	this.chanMsgs = make(chan *LogUnit, C_LOG_CSIZE)
	this.chanExit = make(chan int)

	// threshold
	defer func() {
		threshold, thresholdVal := fmt.Sprintf(C_TH_CHAN_OVERLOAD, this.fileName), C_TH_CHAN_OVERLOAD_VALUE
		if Threshold(threshold) != nil {
			panic("logger instance was existed")
		}
		RegThreshold(threshold, int64(thresholdVal), 10*time.Minute, "ulog internal msg channel overload")
	}()
	if conf == nil {
		return this
	}

	// custom
	if conf.OutMode != 0 {
		this.outMode = conf.OutMode
	}
	this.level = conf.Level
	if conf.DirName != "" {
		if strings.HasPrefix(conf.DirName, "./") {
			this.dirName = conf.DirName
		} else {
			_path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			this.dirName = path.Join(_path, conf.DirName)
		}
	}
	if conf.FileName != "" {
		this.fileName = conf.FileName
	}
	if conf.FileSuffix != "" {
		this.fileSuffix = conf.FileSuffix
	}
	if conf.RotateMax > 0 {
		this.rotateMax = conf.RotateMax
	}
	if conf.RotateSize > 1024 {
		this.rotateSize = conf.RotateSize
	}
	return this
}
func (this *logger) Start() *logger {
	if this.status == ELS_Running {
		return this
	}

	if this.outMode&ELM_File != 0 {
		if err := os.MkdirAll(this.dirName, 0777); err != nil {
			panic(err)
		}

		path := fmt.Sprintf("%s/%s_%s.%s", this.dirName, this.fileName, "system", this.fileSuffix)

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			panic(err)
		}
		this.fileSystmHandle = file
		this.fileSystmLogger = log.New(file, "", log.LstdFlags)
		this.fileSystmLogger.Println("👌")

		this.fileLogicUpdate()
		this.fileLogiclogger.Println("·································START·································")
		this.fileLogiclogger.Println() // add space line
	}

	go this.loop()
	this.status = ELS_Running
	return this
}
func (this *logger) Stop() {
	if this.status != ELS_Running {
		return
	}
	this.chanExit <- 1
	time.Sleep(time.Millisecond * 10)
	this.status = ELS_Exiting
	close(this.chanMsgs)
	this.wgExit.Wait()
}

func (this *logger) GetLevel() ELogLevel { return this.level }

func (this *logger) SetLevel(lv ELogLevel) { this.level = lv }

func (this *logger) UpdPrefix(lvPrefix [ELL_Max]string) {
	this.levelPrefixNames = lvPrefix
}
func (this *logger) AddFilter(filter func(msg *LogUnit) bool) {
	this.filters = append(this.filters, filter)
}

func (this *logger) Trace(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	if !this.canLog(ELL_Trace) {
		return
	}
	this.push(ELL_Trace, depth, fields, fmt.Sprintf(format+"\n", v...))
}
func (this *logger) Tracev(depth int, fields map[string]interface{}, v ...interface{}) {
	if !this.canLog(ELL_Trace) {
		return
	}
	this.push(ELL_Trace, depth, fields, fmt.Sprintln(v...))
}

func (this *logger) Debug(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	if !this.canLog(ELL_Debug) {
		return
	}
	this.push(ELL_Debug, depth, fields, fmt.Sprintf(format+"\n", v...))
}
func (this *logger) Debugv(depth int, fields map[string]interface{}, v ...interface{}) {
	if !this.canLog(ELL_Debug) {
		return
	}
	this.push(ELL_Debug, depth, fields, fmt.Sprintln(v...))
}

func (this *logger) Info(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	if !this.canLog(ELL_Infos) {
		return
	}
	this.push(ELL_Infos, depth, fields, fmt.Sprintf(format+"\n", v...))
}
func (this *logger) Infov(depth int, fields map[string]interface{}, v ...interface{}) {
	if !this.canLog(ELL_Infos) {
		return
	}
	this.push(ELL_Infos, depth, fields, fmt.Sprintln(v...))
}

func (this *logger) Warn(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	if !this.canLog(ELL_Warns) {
		return
	}
	this.push(ELL_Warns, depth, fields, fmt.Sprintf(format+"\n", v...))
}
func (this *logger) Warnv(depth int, fields map[string]interface{}, v ...interface{}) {
	if !this.canLog(ELL_Warns) {
		return
	}
	this.push(ELL_Warns, depth, fields, fmt.Sprintln(v...))
}

func (this *logger) Error(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	this.push(ELL_Error, depth, fields, fmt.Sprintf(format+"\n", v...)) // 输出到队列

	if this.fileSystmHandle != nil {
		this.fileSystmLogger.Print(this.levelPrefixNames[ELL_Error] + " " + fmt.Sprintf(format+"\n", v...)) //直接输出到文件
	}
}
func (this *logger) Errorv(depth int, fields map[string]interface{}, v ...interface{}) {
	this.push(ELL_Error, depth, fields, fmt.Sprintln(v...))

	if this.fileSystmHandle != nil {
		this.fileSystmLogger.Print(this.levelPrefixNames[ELL_Error] + " " + fmt.Sprintln(v...)) //直接输出到文件
	}
}

func (this *logger) Fatal(depth int, fields map[string]interface{}, format string, v ...interface{}) {
	this.push(ELL_Fatal, depth, fields, fmt.Sprintf(format+"\n", v...))

	if this.fileSystmHandle != nil {
		this.fileSystmLogger.Print(this.levelPrefixNames[ELL_Fatal] + " " + fmt.Sprintf(format+"\n", v...)) //直接输出到文件
		this.fileSystmHandle.Sync()
	}
	this.Stop()
	os.Exit(1)
}
func (this *logger) Fatalv(depth int, fields map[string]interface{}, v ...interface{}) {
	if len(v) == 0 || v[0] == nil {
		return
	}
	this.push(ELL_Fatal, depth, fields, fmt.Sprintln(v...))

	if this.fileSystmHandle != nil {
		this.fileSystmLogger.Print(this.levelPrefixNames[ELL_Fatal] + " " + fmt.Sprintln(v...)) //直接输出到文件
		this.fileSystmHandle.Sync()
	}
	this.Stop()
	os.Exit(1)
}

// --------------- Internal logic
func (this *logger) push(level ELogLevel, depth int, fields map[string]interface{}, msg string) {
	if this.status != ELS_Running {
		return
	}
	unit := &LogUnit{Lv: level, At: time.Now(), Fields: fields}
	if depth > 0 {
		file, line, fun := stack(depth)
		unit.Str = fmt.Sprintf("%s %s:%d|%s() %s", level.String(), file, line, fun, msg)
	} else {
		unit.Str = fmt.Sprintf("%s %s", level.String(), msg)
	}
	this.chanMsgs <- unit
	Threshold(fmt.Sprintf(C_TH_CHAN_OVERLOAD, this.fileName)).Assert(int64(len(this.chanMsgs)))
}
func (this *logger) canLog(lev ELogLevel) bool {
	if this.status != ELS_Running {
		return false
	}
	if lev >= this.level {
		return true
	}
	return false
}
func (this *logger) filter(msg *LogUnit) bool {
	if len(this.filters) == 0 {
		return false
	}

	pass := false
	for _, f := range this.filters {
		if f(msg) {
			pass = true
		}
	}
	return pass
}
func (this *logger) fileLogicUpdate() {
	curTime := time.Now()
	if curTime.Year() == this.lastUpdateTime.Year() && curTime.Day() == this.lastUpdateTime.Day() {
		return
	}
	this.lastUpdateTime = curTime

	if this.fileLogicHandle != nil {
		this.fileLogicHandle.Close()
	}

	strTime := fmt.Sprintf("%04d-%02d-%02d", curTime.Year(), curTime.Month(), curTime.Day())

	path := fmt.Sprintf("%s/%s_%s.%s", this.dirName, this.fileName, strTime, this.fileSuffix)
	handle, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	this.fileLogicHandle = handle
	this.fileLogiclogger = log.New(handle, "", 0 /*log.Lmicroseconds*/)
}
func (this *logger) needRenameFile() bool {
	if this.rotateMax > 1 {
		if info, err := this.fileLogicHandle.Stat(); err == nil {
			return info.Size() >= int64(this.rotateSize)
		}
	}
	return false
}
func (this *logger) renameFile() {

	if this.fileLogicHandle != nil {
		this.fileLogicHandle.Close()
	}

	strTime := fmt.Sprintf("%04d-%02d-%02d", this.lastUpdateTime.Year(), this.lastUpdateTime.Month(), this.lastUpdateTime.Day())

	pathmax := fmt.Sprintf("%s/%s_%s.%d.%s", this.dirName, this.fileName, strTime, this.rotateMax, this.fileSuffix)

	if _, err := os.Stat(pathmax); err == nil || os.IsExist(err) {
		os.Remove(pathmax)
	}

	for index := this.rotateMax - 1; index > 0; index-- {
		pathOld := fmt.Sprintf("%s/%s_%s.%d.%s", this.dirName, this.fileName, strTime, index, this.fileSuffix)
		pathNew := fmt.Sprintf("%s/%s_%s.%d.%s", this.dirName, this.fileName, strTime, index+1, this.fileSuffix)
		if _, err := os.Stat(pathOld); err == nil || os.IsExist(err) {
			os.Rename(pathOld, pathNew)
		}
	}

	pathNow := fmt.Sprintf("%s/%s_%s.%s", this.dirName, this.fileName, strTime, this.fileSuffix)
	pathOne := fmt.Sprintf("%s/%s_%s.%d.%s", this.dirName, this.fileName, strTime, 1, this.fileSuffix)
	os.Rename(pathNow, pathOne)
	file, err := os.OpenFile(pathNow, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	this.fileLogicHandle = file
	this.fileLogiclogger.SetOutput(file)
}
func (this *logger) loop() {
	this.wgExit.Add(1)
	t := time.NewTicker(time.Second * 1)

	defer func() {
		if this.fileSystmHandle != nil {
			this.fileSystmLogger.Println("✋")
			this.fileSystmHandle.Close()
		}

		if this.fileLogicHandle != nil {
			this.fileLogiclogger.Println("··································END··································")
			this.fileLogicHandle.Close()
		}

		this.status = ELS_Stopped
		if x := recover(); x != nil {
			if this.fileSystmHandle != nil {
				this.fileSystmLogger.Print("caught panic in logger::loop() error:", x)
			}
		}
		this.wgExit.Done()
	}()

	for {
		select {
		case msg := <-this.chanMsgs:
			if this.filter(msg) {
				continue
			}

			strFields := ""
			if len(msg.Fields) > 0 {
				b, _ := json.Marshal(msg.Fields)
				strFields = fmt.Sprintf("%30s_>%s<_\n", "", string(b))
			}

			if this.outMode&ELM_Std != 0 || msg.Lv >= ELL_Infos {
				this.screenLogger.Println(fmt.Sprintf("\x1b[%dm", LOG_MSG_COLORS[msg.Lv]),
					msg.At.Format("2006-01-02 15:04:05.000"), msg.Str, strFields, "\x1b[0m")
			}
			if this.outMode&ELM_File == 0 {
				continue
			}
			this.fileLogicUpdate()

			this.fileLogiclogger.Println(msg.At.Format("15:04:05.000"), msg.Str, strFields)
		case <-t.C:
			if this.fileLogicHandle != nil {
				this.fileLogicHandle.Sync()
				if this.needRenameFile() {
					this.renameFile()
				}
			}
			if this.fileSystmHandle != nil {
				this.fileSystmHandle.Sync()
			}
		case <-this.chanExit:
			for msg := range this.chanMsgs {
				if this.filter(msg) {
					continue
				}
				if this.outMode&ELM_Std != 0 || msg.Lv >= ELL_Infos {
					this.screenLogger.Println(fmt.Sprintf("\x1b[%dm", LOG_MSG_COLORS[msg.Lv]),
						msg.At.Format("2006-01-02 15:04:05.000"), msg.Str, "\x1b[0m")
				}
				if this.outMode&ELM_File != 0 {
					this.fileLogiclogger.Println(msg.At.Format("15:04:05.000") + " " + msg.Str)
				}
			}
			return
		}
	}
}

func stack(depth int) (file string, line int, fun string) {
	_pc, _file, _line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		paths := strings.Split(_file, "/")
		file = strings.Join(paths[max(0, len(paths)-3):], "/")
	}
	f := runtime.FuncForPC(_pc)
	fields := strings.Split(f.Name(), ".")
	line, fun = _line, fields[len(fields)-1]
	return
}

func max(n1, n2 int) int {
	if n1 > n2 {
		return n1
	}
	return n2
}
