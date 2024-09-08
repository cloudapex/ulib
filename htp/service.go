package htp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudapex/ulib/htp/core"
	"github.com/cloudapex/ulib/htp/metactx"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"google.golang.org/protobuf/proto"
)

var autoBind = autoBinding{}

// API服务
func Service(c *gin.Context, s IService) {

	// bind
	err := doBind(c, s)
	if err != nil {
		doRender(c, s, http.StatusOK, RespBindErr(err))
		return
	}

	// handle
	rsp := doHandle(c, s)

	// render
	doRender(c, s, http.StatusOK, rsp)
}

// --------------- internal

// 解析入参到request
func doBind(c *gin.Context, s IService) error {
	// 全都使用ESBind_Auto
	return c.ShouldBindWith(s, autoBind)
}

// api业务处理
func doHandle(c *gin.Context, s IService) Response {
	CtxRequestSet(c, s)

	return s.Handle(metactx.WithCtx(c))
}

// 提交渲染结果response
func doRender(c *gin.Context, s IService, status int, rsp Response) {
	CtxResponseSet(c, &rsp)

	// render
	mode := ESRender_Json
	if m, ok := s.(ISRenderModer); ok {
		mode = m.RenderMode()
	}

	render := MakeRender(mode)
	if render == nil {
		panic(fmt.Errorf("render not found, mode=%v", mode))
	}
	render.Render(c, status, &rsp)
}

// ==================== 绑定器

// > 自动绑定器
type autoBinding struct{}

func (autoBinding) Name() string { return "auto" }
func (autoBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil {
		return fmt.Errorf("invalid request obj")
	}

	// request参数标记(json编码数据)
	request := "request"
	if req.URL.Query().Has(request) {
		val := req.URL.Query().Get(request)
		// 不需要解析
		if val == "" {
			return nil
		}
		// 从body解
		if ok, _ := convertor.ToBool(val); ok {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}
			return json.Unmarshal(data, obj)
		}
		// 直接解
		data, err := base64.RawURLEncoding.DecodeString(val)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, obj)
	}

	// Request参数标记(pb编码数据)
	request = "Request"
	if req.URL.Query().Has(request) {
		val := req.URL.Query().Get(request)
		// 不需要解析
		if val == "" {
			return nil
		}
		// 从body解
		if ok, _ := convertor.ToBool(val); ok {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}
			return proto.Unmarshal(data, obj.(proto.Message))
		}
		// 直接解
		data, err := base64.RawURLEncoding.DecodeString(val)
		if err != nil {
			return err
		}
		return proto.Unmarshal(data, obj.(proto.Message))
	}

	// 取得 url 里面的参数
	oldMethod := req.Method
	defer func() { req.Method = oldMethod }()
	req.Method = http.MethodGet
	b := binding.Default(req.Method, "")
	if err := b.Bind(req, obj); err != nil {
		return err
	}

	// 取得 body 里面的参数 (form & json)
	oldForm, oldPostForm := req.Form, req.PostForm
	defer func() { req.Form, req.PostForm = oldForm, oldPostForm }()
	req.Method, req.Form, req.PostForm = http.MethodPost, nil, nil
	b = binding.Default(req.Method, filterFlags(req.Header.Get(core.C_HTTP_HEAD_CONTENT_TYPE)))
	if err := b.Bind(req, obj); err != nil {
		return err
	}

	return nil
}
func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
