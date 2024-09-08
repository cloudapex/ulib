package mdb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/cloudapex/ulib/ctl"
	"github.com/cloudapex/ulib/util"

	"xorm.io/xorm"
)

var Ctl IContrler // 默认DB控制器

// 安装控制器
func Install(confs []*Config) ctl.IControler {
	c := ctl.Install(&controller{Confs: confs})
	util.Cast(Ctl == nil, func() { Ctl = c.(IContrler) }, nil)
	return Ctl
}

// Connector 获取指定连接器(引擎)
func Connector(dbName string) *xorm.Engine { return Ctl.Use(dbName) }

// ConnDBStats 获取指定db连接统计状态
func ConnDBStats(dbName string) sql.DBStats {
	// https://zhuanlan.zhihu.com/p/135454743
	return Connector(dbName).DB().DB.Stats()
}

// 创建DBTables
func CreateTable(tabs ...IEntity) error {
	for _, entity := range tabs {
		has, err := Connector(entity.DBName(EDB_Maste)).IsTableExist(entity)
		if err != nil {
			return fmt.Errorf("db table exist for %s, err:%v", entity.TableName(), err)
		}

		// 当 tbl 不存在的时候, 才会创建它
		if !has {
			if err := Connector(entity.DBName(EDB_Maste)).Sync(entity); err != nil {
				return fmt.Errorf("db table synch for %s, err:%v", entity.TableName(), err)
			}
		}
	}
	return nil
}

// 清空DBTables
func DeleteTable(tabs ...IEntity) error {
	for _, entity := range tabs {
		if _, err := Connector(entity.DBName(EDB_Maste)).Where("1=1").Delete(entity); err != nil {
			return err
		}
	}
	return nil
}

// 检测表是否存在
func ExistTable(tbl IEntity) (bool, error) {
	if Ctl == nil {
		return false, errors.New("mdb not install")
	}
	return Ctl.Exist(tbl.DBName(EDB_Maste), tbl)
}

// 同步表结构
func SynchTable(tbls ...IEntity) error {
	if Ctl == nil {
		return errors.New("mdb not install")
	}
	for _, entity := range tbls {
		if err := Ctl.Sync(entity.DBName(EDB_Maste), entity); err != nil {
			return err
		}
	}
	return nil
}

// 删除表(包括索引)
func DropTable(tbls ...IEntity) error {
	if Ctl == nil {
		return errors.New("mdb not install")
	}
	for _, entity := range tbls {
		if err := Ctl.Drop(entity.DBName(EDB_Maste), entity); err != nil {
			return err
		}
	}
	return nil
}

// ======================================== ITabler

// 主(Master) Table
func MTable(tbl IEntity) *table { return Table(EDB_Maste, tbl, nil) }

// 从(Slave) Table
func STable(tbl IEntity) *table { return Table(EDB_Slave, tbl, nil) }

// 事务(Transaction) Table
func TTable(tx *Session, tbl IEntity) *table { return Table(EDB_Maste, tbl, tx.Session) }

// ======================================== Transaction

// 创建事务(Transation session)(当不需要真正的事务时可以使用 mdb.TxNil)
func Transation(dbName string) *Session { return &Session{Session: Connector(dbName).NewSession()} }

// ======================================== SQL Operate

// 原生SQL执行(one)[beans:输出, query:语句(支持?), args:语句参数]
func SQLGet(dbName string, beans []interface{}, query interface{}, args ...interface{}) (bool, error) {
	return Connector(dbName).SQL(query, args...).Get(beans...)
}

// 原生SQL执行(multi)[rowsSlicePtr:输出slice或map, query:语句(支持?), args:语句参数]
func SQLFind(dbName string, rowsSlicePtr interface{}, query interface{}, args ...interface{}) error {
	return Connector(dbName).SQL(query, args...).Find(rowsSlicePtr)
}

// 原生SQL执行(page)[rowsSlicePtr:输出slice或map, query:语句(支持?), args:语句参数]
func SQLFindPage(dbName string, rowsSlicePtr interface{}, query interface{}, args ...interface{}) (int64, error) {
	return Connector(dbName).SQL(query, args...).FindAndCount(rowsSlicePtr)
}

// 原生SQL执行(delete|update)[query:语句(支持?), args:语句参数]
func SQLExec(tx *Session, query interface{}, args ...interface{}) (int64, error) {
	sqlOrArgs := []interface{}{query}
	sqlOrArgs = append(sqlOrArgs, args)

	ret, err := tx.Exec(sqlOrArgs...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// 原生SQL执行(delete|update)[query:语句(支持?), args:语句参数]
func SQLExecute(dbName string, query interface{}, args ...interface{}) (int64, error) {
	sqlOrArgs := []interface{}{query}
	sqlOrArgs = append(sqlOrArgs, args)

	ret, err := Connector(dbName).Exec(sqlOrArgs...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}
