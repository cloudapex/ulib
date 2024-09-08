package mdb

import (
	"github.com/cloudapex/ulib/log"

	"xorm.io/xorm"
)

// 事务类型
type Session struct {
	*xorm.Session
	err        error
	defers     []func() // Commit 之后 按顺序执行
	deferBacks []func() // Rollback 之后 按顺序执行
}

// 事务开启和关闭(配合 defer)
func (tx *Session) Begin() error    { return tx.Session.Begin() }
func (tx *Session) Close(err error) { tx.commitRollback(); tx.Session.Close() }

// 设置错误
func (tx *Session) Error(err error) error {
	if tx.Session == nil {
		return err
	}
	tx.err = err
	return err
}

// 添加 DeferFun when Commit
func (tx *Session) Defer(f func()) {
	if tx.Session == nil {
		f()
		return
	}
	tx.defers = append(tx.defers, f)
}

// 添加 DeferBackFun when Rollback
func (tx *Session) DeferBack(f func()) {
	if tx.Session != nil {
		tx.deferBacks = append(tx.deferBacks, f)
	}
}

// 事务提交和回滚(配合 derfer)
func (tx *Session) commitRollback() {
	var ret error
	if tx.err != nil {
		ret = tx.Rollback()
		for _, f := range tx.deferBacks {
			f()
		}
	} else {
		if ret = tx.Commit(); ret == nil {
			for _, f := range tx.defers {
				f()
			}
		}
	}
	if ret != nil {
		log.ErrorD(1, "tx.CommitRollback err:%v", ret)
	}
}
