package util

/******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2014-10-18
Description:系统函数
- 文件操作: https://studygolang.com/articles/3154
******************************************************************************/
import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cloudapex/ulib/log"

	"github.com/duke-git/lancet/v2/mathutil"
	"golang.org/x/exp/rand"
)

var (
	chanSig     = make(chan os.Signal, 1)
	chanExit    = make(chan int)
	termHandles = []func(interface{}){}
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
	signal.Notify(chanSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	Term(func(reason interface{}) { log.Term() })
	go func() {
		c := <-chanSig
		for i := len(termHandles) - 1; i >= 0; i-- {
			termHandles[i](c)
		}

		exitCode := 0
		select {
		case chanExit <- exitCode:
			return
		case <-time.After(time.Second * 1):
			os.Exit(exitCode)
		}
	}()
}

func Wait(x ...interface{}) {
	<-chanExit
}
func Term(clearHand func(interface{})) {
	termHandles = append(termHandles, clearHand)
}
func Quit(delay ...time.Duration) {
	go func() {
		if len(delay) > 0 && delay[0] > 0 {
			<-time.After(delay[0])
		}
		close(chanSig)
	}()
}

func GoId() (int, error) {
	var buf [64]byte
	idField := strings.Fields(strings.TrimPrefix(string(buf[:runtime.Stack(buf[:], false)]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return 0, fmt.Errorf("cannot get goroutine id: %v", err)
	}
	return id, nil
}

func Catch(desc string, x interface{}, bFatal ...bool) bool {
	if x == nil {
		return false
	}
	head := fmt.Sprintf("%s: %v\n", desc, x)

	buf := make([]byte, 256*10)
	size := runtime.Stack(buf, true)
	stack := string(buf[0:size])
	Cast(len(bFatal) > 0 && bFatal[0], func() { log.FatalD(-1, "%s%s", head, stack) }, func() { log.ErrorD(-1, "%s%s", head, stack) })
	return true
}

func Sleep(d time.Duration) { <-time.After(d) }

func Stack(depth int) (file string, line int, fun string) {
	_pc, _file, _line, ok := runtime.Caller(1 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		file = _file
		slash := strings.Split(_file, "/")
		if len(slash) >= 0 {
			file = strings.Join(slash[mathutil.Max(len(slash)-3, 0):], "/")
		}
	}
	f := runtime.FuncForPC(_pc)
	fields := strings.Split(f.Name(), ".")
	line, fun = _line, fields[len(fields)-1]
	return
}

func Goroutine(name string, goFun func(), wg ...*sync.WaitGroup) {
	go func() {
		if w := DefaultVal(wg); w != nil {
			w.Add(1)
			defer func() { w.Done() }()
		}
		defer func() { Catch(fmt.Sprintf("Goroutine[%s] panic", name), recover()) }()
		goFun()
	}()
}

func TimeStart() time.Time {
	return startupAt
}
func TimeLived() time.Duration {
	return time.Since(startupAt)
}

func StructName(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}
func ExeName() string {
	return strings.Split(ExeFullName(), ".")[0]
}

func ExeFullName() string {
	return filepath.Base(os.Args[0])
}

func FuncName(fun interface{}) string {
	return FuncFullName(fun, '.')
}
func FuncFullName(fun interface{}, seps ...rune) string {
	return FuncFullNameRef(reflect.ValueOf(fun), seps...)
}
func FuncFullNameRef(valFun reflect.Value, seps ...rune) string {
	fn := runtime.FuncForPC(valFun.Pointer()).Name()
	if len(seps) == 0 {
		return fn
	}

	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})
	if size := len(fields); size > 0 {
		return strings.Split(fields[size-1], "-")[0]
	}
	return fn
}
