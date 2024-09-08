package mdb

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2020-06-15
Description: 自定义日志接口[xorm]
*******************************************************************************/
import (
	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"strings"

	xlog "xorm.io/xorm/log"
)

// ==================== XormLogger
type XormLogger struct {
	Lv      xlog.LogLevel
	ShowSql bool
} //
func (this *XormLogger) Debug(v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_DEBUG, func() { log.DebugDv(-1, append([]interface{}{"XORM: "}, v...)) }, nil)
}
func (this *XormLogger) Debugf(format string, v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_DEBUG, func() { log.DebugD(-1, "XORM: "+format, v...) }, nil)
}
func (this *XormLogger) Info(v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_INFO, func() { log.InfoDv(-1, append([]interface{}{"XORM: "}, v...)) }, nil)
}
func (this *XormLogger) Infof(format string, v ...interface{}) {
	if strings.Contains(format, "PING ") {
		log.DebugD(-1, "XORM: "+format, v...)
		return
	}
	util.Cast(this.Lv <= xlog.LOG_INFO, func() { log.InfoD(-1, "XORM: "+format, v...) }, nil)
}
func (this *XormLogger) Warn(v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_WARNING, func() { log.WarnDv(-1, append([]interface{}{"XORM: "}, v...)) }, nil)
}
func (this *XormLogger) Warnf(format string, v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_WARNING, func() { log.WarnD(-1, "XORM: "+format, v...) }, nil)
}
func (this *XormLogger) Error(v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_ERR, func() { log.ErrorDv(-1, append([]interface{}{"XORM: "}, v...)) }, nil)
}
func (this *XormLogger) Errorf(format string, v ...interface{}) {
	util.Cast(this.Lv <= xlog.LOG_ERR, func() { log.ErrorD(-1, "XORM: "+format, v...) }, nil)
}
func (this *XormLogger) Level() xlog.LogLevel { return this.Lv }

func (this *XormLogger) SetLevel(l xlog.LogLevel) { this.Lv = l }

func (this *XormLogger) ShowSQL(show ...bool) {
	util.Cast(len(show) == 0, func() { this.ShowSql = true }, func() { this.ShowSql = show[0] })
}
func (this *XormLogger) IsShowSQL() bool { return this.ShowSql }
