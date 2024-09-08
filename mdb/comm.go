package mdb

import (
	"github.com/cloudapex/ulib/ctl"

	"xorm.io/core"
	"xorm.io/xorm"
)

// 控制器接口
type IContrler interface {
	ctl.IControler

	// 使用指定DB引擎
	Use(name string) *xorm.Engine

	// 表存在不存在
	Exist(name string, tbl IEntity) (bool, error)

	// 同步表结构
	Sync(name string, tbls ...IEntity) error

	// 移除表
	Drop(name string, tbls ...IEntity) error
}

// > 实体接口
type IEntity interface {

	// 连接名
	DBName(e EDB) string

	// 表名
	TableName() string
}

// > 查询条件选项
type CondOpterFunc func(s *xorm.Session) *xorm.Session

func (f CondOpterFunc) ApplyOpt(s *xorm.Session) *xorm.Session { return f(s) }

type CondOpter interface {
	ApplyOpt(s *xorm.Session) *xorm.Session
}

// > DB表字段字节数据转换接口
type IFieldData = core.Conversion
