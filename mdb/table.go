package mdb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudapex/ulib/util"

	"github.com/duke-git/lancet/v2/mathutil"
	"xorm.io/xorm"
)

func Table(e EDB, tbl IEntity, tx ...*xorm.Session) *table {
	var s *xorm.Session
	util.Cast(len(tx) > 0, func() { s = tx[0] }, nil)
	return &table{tbl, e, s}
}

// > SQL Table
type table struct {
	IEntity
	edb   EDB
	sessn *xorm.Session
}

// 加载指定记录(one) (this:即作为条件又作为结果)
func (this *table) Load(opts ...CondOpter) (bool, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Get(this.IEntity)
}

// 检测记录存在 (this:作为条件)
func (this *table) Exist(opts ...CondOpter) (bool, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Exist(this.IEntity)
}

// 统计记录数量 (this:作为条件)
func (this *table) Count(opts ...CondOpter) (int64, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Count(this.IEntity)
}

// 添加记录 (others为空时:添加this; 否则:批量添加others)(注意默认值)
func (this *table) Create(others ...IEntity) (int64, error) {
	t := this.session()
	if len(others) == 0 {
		return t.Insert(this.IEntity)
	}
	return t.Insert(others)
}

// 更新记录(按需) (this:作为条件,upd:新记录)
func (this *table) Update(upd IEntity, opts ...CondOpter) (int64, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Update(upd, this.IEntity)
}

// 更新记录(字段) (this:作为条件,upd:新记录,eFields:强制更新的字段)
func (this *table) UpdFields(upd IEntity, fields []string, opts ...CondOpter) (int64, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.MustCols(fields...).Update(upd, this.IEntity)
}

// 删除记录 (this:作为条件)
func (this *table) Delete(opts ...CondOpter) (int64, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Delete(this.IEntity)
}

// 查询模型记录(one) (this:作为条件, bean:输出ptr)
func (this *table) Get(bean interface{}, opts ...CondOpter) (bool, error) {
	t := this.session()
	t = this.apply(t, opts)
	return t.Get(bean)
}

// 查询模型字段(one) (this:作为条件, beans:输出ptr(&user or &id &name ... or map or slice))
func (this *table) GetFields(beans []interface{}, fields []string, opts ...CondOpter) (bool, error) {
	if len(fields) == 0 {
		return false, ErrSelectFieldsNone
	}

	t := this.session()
	t = t.Cols(fields...)
	t = this.apply(t, opts)
	return t.Get(beans...)
}

// 查询模型记录(all) (this:作为条件, rowsSlicePtr:输出slice或map)
func (this *table) Find(rowsSlicePtr interface{}, opts ...CondOpter) error {
	t := this.session()
	t = this.apply(t, opts)
	return t.Find(rowsSlicePtr, this.IEntity)
}

// 查询模型记录(page) (this:作为条件, rowsSlicePtr:输出slice或map)
func (this *table) FindPage(rowsSlicePtr interface{}, page, pnum int, opts ...CondOpter) (int64, error) {
	page, pnum = mathutil.Max(1, page), mathutil.Max(1, pnum)

	opts = append(opts, COPage(page, pnum))

	t := this.session()
	t = this.apply(t, opts)
	return t.FindAndCount(rowsSlicePtr, this.IEntity)
}

// 查询模型字段(all) (this:作为条件, rowsSlicePtr:输出slice或map, fields:取哪些列)
func (this *table) FindFields(rowsSlicePtr interface{}, fields []string, opts ...CondOpter) error {
	if len(fields) == 0 {
		return ErrSelectFieldsNone
	}

	t := this.session()
	t = t.Cols(fields...)
	t = this.apply(t, opts)
	return t.Find(rowsSlicePtr, this.IEntity)
}

// 查询模型字段(page) (this:作为条件, rowsSlicePtr:输出slice或map, fields:取哪些列)
func (this *table) FindFieldsPage(rowsSlicePtr interface{}, fields []string, page, pnum int, opts ...CondOpter) (int64, error) {
	if len(fields) == 0 {
		return 0, ErrSelectFieldsNone
	}

	page, pnum = mathutil.Max(1, page), mathutil.Max(1, pnum)

	opts = append(opts, COPage(page, pnum))

	t := this.session()
	t = t.Cols(fields...)
	t = this.apply(t, opts)
	return t.FindAndCount(rowsSlicePtr, this.IEntity)
}

// 查询自定义字段 (this:作为条件, rowsSlicePtr:输出slice或map或ptr, fields:为select中的自定义字段','分割)
func (this *table) Select(rowsSlicePtr interface{}, fields string, opts ...CondOpter) error {
	if len(fields) == 0 {
		return ErrSelectFieldsNone
	}

	t := this.session()
	t = t.Select(fields)
	t = this.apply(t, opts)
	return t.Find(rowsSlicePtr, this)
}

// 查询自定义字段(page) (this:作为条件, rowsSlicePtr:输出slice或map, fields:为select中的自定义字段','分割)
func (this *table) SelectPage(rowsSlicePtr interface{}, fields string, page, pnum int, opts ...CondOpter) (int64, error) {
	if len(fields) == 0 {
		return 0, ErrSelectFieldsNone
	}
	page, pnum = mathutil.Max(1, page), mathutil.Max(1, pnum)

	opts = append(opts, COPage(page, pnum))

	t := this.session()
	t = t.Select(fields)
	t = this.apply(t, opts)
	return t.FindAndCount(rowsSlicePtr, this)
}

// 查询自定义联合字段 (this:不作条件, rowsSlicePtr:输出slice或map或ptr, fields:为select中的自定义字段','分割)
func (this *table) SelectJoin(rowsSlicePtr interface{}, fields string, joins []*Join, opts ...CondOpter) error {
	if len(fields) == 0 {
		return ErrSelectFieldsNone
	}

	t := this.session()
	t = t.Select(fields)
	for _, jo := range joins {
		t = t.Join(jo.Type, jo.Tablename, jo.OnCond)
	}
	t = this.apply(t, opts)
	return t.Find(rowsSlicePtr)
}

// 查询自定义联合字段(page) (this:不作条件, rowsSlicePtr:输出slice或map, fields:为select中的自定义字段','分割)
func (this *table) SelectJoinPage(rowsSlicePtr interface{}, fields string, joins []*Join, page, pnum int, opts ...CondOpter) (int64, error) {
	if len(fields) == 0 {
		return 0, ErrSelectFieldsNone
	}
	page, pnum = mathutil.Max(1, page), mathutil.Max(1, pnum)

	opts = append(opts, COPage(page, pnum))

	t := this.session()
	t = t.Select(fields)
	for _, jo := range joins {
		t = t.Join(jo.Type, jo.Tablename, jo.OnCond)
	}
	t = this.apply(t, opts)
	return t.FindAndCount(rowsSlicePtr)
}

// 应用选项
func (this *table) apply(s *xorm.Session, opts []CondOpter) *xorm.Session {
	for _, o := range opts {
		s = o.ApplyOpt(s)
	}
	return s
}

// supply Session
func (this *table) session() *xorm.Session {
	if this.sessn != nil {
		return this.sessn
	}
	return Connector(this.DBName(this.edb)).Table(this.IEntity)
}

// ==================== 查询选项

// 查询选项[SQL]-自定义SQL条件(支持?)
func COWhere(query string, args ...interface{}) CondOpterFunc {
	return func(s *xorm.Session) *xorm.Session { s = s.Where(query, args...); return s }
}

// 查询选项[SQL]-自定义SQL条件(支持?)
func COAndOr(e EAndOr, query string, args ...interface{}) CondOpterFunc {
	if e == EAO_And {
		return func(s *xorm.Session) *xorm.Session { s = s.And(query, args...); return s }
	}
	return func(s *xorm.Session) *xorm.Session { s = s.Or(query, args...); return s }
}

// 查询选项[ORM]-匹配
func COLike(field string, val string) CondOpterFunc {
	return func(s *xorm.Session) *xorm.Session { s = s.Where(fmt.Sprintf("%s like ?", field), val); return s }
}

// 查询选项[ORM]-比较
func COCompare(field string, e ECompare, val interface{}) CondOpterFunc {
	return func(s *xorm.Session) *xorm.Session { s = s.Where(fmt.Sprintf("%s %s ?", field, e), val); return s }
}

// 查询选项[ORM]-集合
func COInNotIn(field string, e EContain, args ...interface{}) CondOpterFunc {
	if len(args) == 0 {
		return func(s *xorm.Session) *xorm.Session { return s }
	}
	v := reflect.ValueOf(args[0])
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return func(s *xorm.Session) *xorm.Session { return s }
		}
	}
	if e == EContain_In {
		return func(s *xorm.Session) *xorm.Session { s = s.In(field, args...); return s }
	}
	return func(s *xorm.Session) *xorm.Session { s = s.NotIn(field, args...); return s }
}

// 查询选项[ORM]-分组
func COGroup(fields ...string) CondOpterFunc {
	if len(fields) == 0 {
		return func(s *xorm.Session) *xorm.Session { return s }
	}
	return func(s *xorm.Session) *xorm.Session { s = s.GroupBy(strings.Join(fields, ",")); return s }
}

// 查询选项[ORM]-过滤 (selAlias:select 中的聚合别名)
func COHaving(selAlias string, e ECompare, val interface{}) CondOpterFunc {
	return func(s *xorm.Session) *xorm.Session { s = s.Having(fmt.Sprintf("%s %s %v", selAlias, e, val)); return s }
}

// 查询选项[ORM]-排序
func COOrder(e ESort, fields ...string) CondOpterFunc {
	if len(fields) == 0 {
		return func(s *xorm.Session) *xorm.Session { return s }
	}
	return func(s *xorm.Session) *xorm.Session {
		if e == ESort_Asc {
			s = s.Asc(fields...)
		} else {
			s = s.Desc(fields...)
		}
		return s
	}
}

// 查询选项[ORM]-分页
func COPage(page, pnum int) CondOpterFunc {
	page--
	return func(s *xorm.Session) *xorm.Session { s = s.Limit(pnum, page*pnum); return s }
}

// 查询选项[ORM]-限量
func COLimit(num int) CondOpterFunc {
	return func(s *xorm.Session) *xorm.Session { s = s.Limit(num); return s }
}

// --------------- Join
/* Join 语法自由度过高,不适合封装,按照下面格式自己玩
type UserDetail struct {
    User `xorm:"extends"`
    Detail `xorm:"extends"`
}

var users []UserDetail
err := engine.Table("user").Select("user.*, detail.*").
    Join("INNER", "detail", "detail.user_id = user.id").
	Where("user.name = ?", name).
	Limit(10, 0).Find(&users)

SQL Out: SELECT user.*, detail.* FROM user INNER JOIN detail WHERE user.name = ? limit 10 offset 0
*/

// --------------- ON DUPLICATE KEY
/* 插入并更新语句 不适合封装, 按照下面格式自己玩
INSERT INTO daily_sign(user_id,continue_days,sign_date,udid) VALUES(?,?,?,?)  ON DUPLICATE KEY UPDATE continue_days=?,sign_date=?,udid=?
*/
//------------------------------------------------------------------------------

// > 样例
// type TableBase struct {
// 	Id int64
// }

// func (this *TableBase) DBName(e EDB) string {
// 	return EDBToName(e, "xx-master", "xx-slave")
// }

// func (this *TableBase) TableName() string { return "table_base" }
