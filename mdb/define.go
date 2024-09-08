package mdb

import (
	"errors"
	"fmt"
)

var (
	TxNil = &Session{} // 空事务对象

	ErrNameRepeated     = errors.New("ErrMdbNameRepeated")
	ErrSelectFieldsNone = errors.New("ErrSelectFieldsZero")
)

// > DB配置
type Config struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Store   string `json:"store"`
	User    string `json:"user"`
	Passwd  string `json:"passwd"`
	IdleMax int64  `json:"idle_max"`
	OpenMax int64  `json:"open_max"`
	ShowSql bool   `json:"show_sql"`
}

// > Engine类型
type EDB int //
const (
	EDB_Maste EDB = iota + 1 // 主库
	EDB_Slave                // 从库
)

// > join
type Join struct {
	Type      string      // INNER(default)|LEFT JOIN|RIGHT JOIN
	Tablename interface{} // Tablename
	OnCond    string      //  eg: on detail.user_id = user.id
}

// > 排序类型
type ESort int //
const (
	ESort_Asc  ESort = iota + 1 // 正序 从小到大
	ESort_Desc                  // 倒序 从大到小
)

// > 比较类型
type ECompare int //
const (
	ECompare_LT   ECompare = iota + 1 // <  value
	ECompare_LTE                      // <= value
	ECompare_GT                       // >  value
	ECompare_GTE                      // >= value
	ECompare_NotE                     // != value
	ECompare_EQ                       // = value
) //
func (e ECompare) String() string {
	switch e {
	case ECompare_LT:
		return "<"
	case ECompare_LTE:
		return "<="
	case ECompare_GT:
		return ">"
	case ECompare_GTE:
		return ">="
	case ECompare_NotE:
		return "!="
	case ECompare_EQ:
		return "="
	}
	return fmt.Sprintf("ECompare_Unkonw(%d)", e)
}

// > 集合包含
type EContain int //
const (
	EContain_In    EContain = iota + 1 // in
	EContain_NotIn                     // not in
)

// > 并且或者
type EAndOr int //
const (
	EAO_And EAndOr = iota + 1 // and
	EAO_Or                    // or
)
