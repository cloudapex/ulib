package htp

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudapex/ulib/log"
	"github.com/cloudapex/ulib/util"

	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Controller(conf *Config) IContrler { return &controller{Conf: conf} }

// > htp control
type controller struct {
	ser http.Server

	groups []IGroupRouter

	Conf *Config
}

func (this *controller) HandleName() string { return "htp" }

func (this *controller) HandleInit() {

	util.Cast(this.Conf == nil, func() { log.Fatal("conf = nil") }, nil)

	util.Cast(len(units) != 0, func() { this.groups = append(this.groups, units...) }, nil)

	gin.SetMode(this.Conf.RunMode)
	gin.DefaultWriter, gin.DefaultErrorWriter = &GinLogger{}, &GinRecover{}

	this.initRouter()
	this.startServer()
}
func (this *controller) HandleTerm() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := this.ser.Shutdown(ctx); err != nil {
		log.Error("shutdown err:%v", err)
	}
}

// ==================== internal

func (this *controller) initRouter() {
	log.TraceD(-1, "Start init GroupRouter(%d)...", len(this.groups))
	defer log.InfoD(-1, "Init GroupRouter(%d) done.", len(this.groups))

	r := gin.New()
	util.Cast(this.Conf.RunMode == "debug", func() { r.Use(gin.Logger()) }, nil)
	r.Use(gin.Recovery())

	this.ser.Handler = h2c.NewHandler(r, &http2.Server{})

	// 1. init root router
	for _, it := range this.groups {
		if it.Name() == "" {
			log.Fatal("routerGroup.Name() is empty. type:%#v", it)
		}
		if strutil.ContainsAny(it.Name(), []string{".", "/"}) { // match root router
			it.Init(&GroupRouter{RouterGroup: &r.RouterGroup})
		}
	}

	// 1. init other router
	existeds := map[string]*GroupRouter{}
	for _, it := range this.groups {
		if it.Name() == "" {
			log.Fatal("routerGroup.Name() is empty. type:%#v", it)
		}

		if strutil.ContainsAny(it.Name(), []string{".", "/"}) { // match root router
			continue
		}

		if _, ok := existeds[it.Name()]; !ok {
			_g := r.Group(it.Name())
			existeds[it.Name()] = &GroupRouter{RouterGroup: _g}
		}
		it.Init(existeds[it.Name()])
	}
}
func (this *controller) startServer() {
	log.TraceD(-1, "Start htp server...")
	defer log.InfoD(-1, "Start htp server on listen addr:%q", this.Conf.ListenAddr)

	this.ser.Addr = this.Conf.ListenAddr
	this.ser.ReadHeaderTimeout = 2 * time.Second // 读取请求头超时时间
	this.ser.IdleTimeout = 60 * time.Second      // 连接的空闲超时时间
	this.ser.MaxHeaderBytes = 4 * 1024           // 请求和回应标头的最大大小为 1 MB
	this.ser.ReadTimeout = time.Duration(mathutil.Max(5, this.Conf.ReadTimeout)) * time.Second
	this.ser.WriteTimeout = time.Duration(mathutil.Max(10, this.Conf.WriteTimeout)) * time.Second
	this.ser.SetKeepAlivesEnabled(true)

	if !this.Conf.ListnTls.Enable {
		go func() {
			if err := this.ser.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("Gin run err:%v", err)
			}
		}()
	} else {
		go func() {
			err := this.ser.ListenAndServeTLS(this.Conf.ListnTls.CrtFile, this.Conf.ListnTls.KeyFile)
			if err != nil && err != http.ErrServerClosed {
				log.Fatal("Gin run with TLS err:%v", err)
			}
		}()
	}
}
