package mdb

import (
	"fmt"

	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/duke-git/lancet/v2/mathutil"
	_ "github.com/go-sql-driver/mysql"

	"xorm.io/core"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

func Controller() IContrler { return &controller{} }

// > mysql-client control
type controller struct {
	mapEngines map[string]*xorm.Engine

	Confs []*Config
}

func (this *controller) HandleName() string { return "mdb" }

func (this *controller) HandleInit() {
	util.Cast(this.Confs == nil, func() { log.Fatal("conf = nil") }, nil)

	this.mapEngines = make(map[string]*xorm.Engine)

	log.TraceD(-1, "Start init mysql connect(%d)...", len(this.Confs))
	defer log.InfoD(-1, "Init mysql connect(%d) done.", len(this.Confs))
	for _, conf := range this.Confs {
		if _, ok := this.mapEngines[conf.Name]; ok {
			log.Fatal("init err=%v", ErrNameRepeated)
		}
		hand, err := newHand(conf)
		if err != nil {
			log.Fatal("init err=%v", err)
		}
		this.mapEngines[conf.Name] = hand
	}
}
func (this *controller) HandleTerm() {
	for _, h := range this.mapEngines {
		h.Close()
	}
}

//  ==================== Functions
func (this *controller) Use(name string) *xorm.Engine {
	return this.mapEngines[name]
}
func (this *controller) Exist(name string, tbl IEntity) (bool, error) {
	return this.Use(name).IsTableExist(tbl)
}
func (this *controller) Sync(name string, tbls ...IEntity) error {
	return this.Use(name).Sync(tbls)
}
func (this *controller) Drop(name string, tbls ...IEntity) error {
	return this.Use(name).DropTables(tbls)
}

// ------------------------------------------------------------------------------
func newHand(conf *Config) (*xorm.Engine, error) {
	// conf.revise()

	source := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.User, conf.Passwd, conf.Host, conf.Store)
	x, err := xorm.NewEngine("mysql", source)
	if err != nil {
		return nil, err
	}

	x.SetLogger(&XormLogger{})
	x.SetLogLevel(xlog.LogLevel(mathutil.Max(int(log.GetLevel()-1), 0)))

	if err := x.Ping(); err != nil {
		return nil, err
	}

	x.SetTableMapper(core.SnakeMapper{})
	x.SetColumnMapper(core.SnakeMapper{})
	x.SetMaxIdleConns(int(conf.IdleMax))
	x.SetMaxOpenConns(int(conf.OpenMax))
	//x.SetTableMapper(core.NewPrefixMapper(core.SnakeMapper{}, C_TABLE_PREFIX))

	x.ShowSQL(conf.ShowSql)

	return x, nil
}
