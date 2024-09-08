package htp

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.2
Date:2014-10-18
Description: 接口定义
*******************************************************************************/
import (
	"github.com/cloudapex/ulib/ctl"
	"github.com/cloudapex/ulib/htp/metactx"

	"github.com/gin-gonic/gin"
)

// > 控制器接口
type IContrler interface {
	ctl.IControler
}

// > 路由接口
type IGroupRouter interface {

	// 模块名称
	Name() string // 返回["", ".", "/"], 则表示 root router

	// 初始化路由
	Init(group *GroupRouter)
}

// > API注册观察回调函数
type TObserveCallBack func(path string, service IService)

// > render接口
type IRender interface {

	// 模式
	Mode() ESRenderMode

	// 渲染
	Render(c *gin.Context, status int, rsp *Response)
}

// > API服务接口

// API服务接口(基本)
type IService interface {
	Handle(meta metactx.IContext) Response // 处理器
}

// API服务接口+(渲染模式)
type ISRenderModer interface {
	RenderMode() ESRenderMode
}
