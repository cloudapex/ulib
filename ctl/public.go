package ctl

import (
	"flag"
	"fmt"

	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/spf13/pflag"
)

var (
	app      appInfo
	controls = []IControler{}
)

func init() {
	util.Term(func(reason interface{}) { shut(fmt.Errorf("%v", reason)) })
}

// 初始化
func Init(name, version string, conf ...*log.Config) interface{} {
	app.Name, app.Version = name, version
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if len(conf) > 0 && conf[0] != nil {
		log.Init(conf[0])
	} else {
		log.Init(nil)
	}
	return nil
}

// 等待结束
func Wait(x interface{}) {
	log.InfoD(-1, "[app_name:%q exe_name:%q app_ver:%s] Start Work...", AppName(), util.ExeName(), AppVersion())

	util.Wait()
}

// 终止运行
func Shut(x interface{}) {
	util.Quit()
	util.Wait()
}

// ======================================== [control]
// 安装控制器
func Install(ctrl IControler) IControler {
	if Controler(ctrl.HandleName()) != nil {
		log.Fatal("Control[%v] was already existed.", ctrl.HandleName())
	}

	ctrl.HandleInit()
	controls = append(controls, ctrl)
	return ctrl
}

// 获取控制器
func Controler(name string) IControler {
	for _, it := range controls {
		if it.HandleName() == name {
			return it
		}
	}
	return nil
}

// ======================================== [function]

// AppName
func AppName() string { return app.Name }

// AppVersion
func AppVersion() string { return app.Version }

// ctrl field logger
func Logger(name string) log.ILoger { return log.Field("ctrl", name) }

// ======================================== [internal]

// 停止所有控制器(同步)
func shut(reason error) {
	log.InfoD(-1, "Start Shut Controls(%d)... by reason: %q", len(controls), reason)
	defer log.InfoD(-1, "[app_name:%q exe_name:%q app_ver:%s] Shut Done.", AppName(), util.ExeName(), AppVersion())

	for i := len(controls) - 1; i >= 0; i-- {
		controls[i].HandleTerm()
	}
}
