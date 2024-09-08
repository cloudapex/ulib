package rdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

func Controller(confs []*Config) IContrler { return &controller{Confs: confs} }

// > redis-client controller
type controller struct {
	pools map[string]IPooler

	Confs []*Config
}

func (this *controller) HandleName() string { return "rdb" }

func (this *controller) HandleInit() {

	util.Cast(len(this.Confs) == 0, func() { log.Fatal("conf = nil") }, nil)

	this.pools = make(map[string]IPooler)

	log.TraceD(-1, "Start init redis conn pool(%d)...", len(this.Confs))
	defer log.InfoD(-1, "Init redis conn pool(%d) done.", len(this.Confs))

	for _, conf := range this.Confs {
		if _, ok := this.pools[conf.Name]; ok {
			log.Fatalv(ErrNameRepeated)
		}
		// conf.revise()
		var err error
		var pool IPooler
		if conf.Name == C_DB_CLUSTER {
			pool, err = newClusterPool(conf)
		} else {
			pool, err = newNormalPool(conf)
		}
		if err != nil {
			log.Fatalv(err)
		}

		this.pools[conf.Name] = pool
	}
}

func (this *controller) HandleTerm() {
	for _, p := range this.pools {
		p.Close()
	}
}

//  ==================== Functions

// Use 选择连接池
func (this *controller) Use(name string) IPooler {
	if it, ok := this.pools[C_DB_CLUSTER]; ok {
		return it
	}
	return this.pools[name]
}

// ------------------------------------------------------------------------------
func newNormalPool(conf *Config) (IPooler, error) {
	p := &NormalPool{redis.Pool{
		MaxIdle:     int(conf.MaxIdle),
		MaxActive:   int(conf.MaxActive),
		IdleTimeout: C_CONN_IDLE_TIMEOVER,
		Wait:        conf.Wait,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", conf.Addr)
			if err != nil {
				return nil, fmt.Errorf("dail err:%v", err)
			}
			if len(conf.Passwd) != 0 {
				if _, err := conn.Do("AUTH", conf.Passwd); err != nil {
					conn.Close()
					return nil, fmt.Errorf("Do('AUTH') err:%v", err)
				}
			}
			if _, err := conn.Do("SELECT", conf.DbIdx); err != nil {
				conn.Close()
				return nil, fmt.Errorf("Do('SELECT') err:%v", err)
			}
			// r, err := redis.String(conn.Do("PING"))
			// if err != nil || r != "PONG" {
			// 	return nil, errors.Errorf("Do('PING') err=%v ret=%v", err, r)
			// }
			return conn, nil
		},
	}}

	c := p.Get()
	if err := c.Err(); err != nil {
		return nil, err
	}
	defer c.Close()

	return p, nil
}
func newClusterPool(conf *Config) (IPooler, error) {
	opts := []redis.DialOption{redis.DialConnectTimeout(5 * time.Second)}
	if len(conf.Passwd) != 0 {
		opts = append(opts, redis.DialPassword(conf.Passwd))
	}

	p := &ClusterPool{redisc.Cluster{
		StartupNodes: strings.Split(conf.Addr, ","),
		DialOptions:  opts,
		CreatePool: func(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
			return &redis.Pool{
				MaxIdle:     int(conf.MaxIdle),
				MaxActive:   int(conf.MaxActive),
				IdleTimeout: time.Minute,
				Wait:        conf.Wait,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("tcp", addr, opts...)
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}, nil
		},
	}}
	if err := p.Refresh(); err != nil {
		return nil, fmt.Errorf("refresh failed: %v", err)
	}
	return p, nil
}
