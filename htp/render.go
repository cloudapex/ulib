package htp

import (
	"github.com/cloudapex/ulib/htp/pb"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var renders []IRender

func init() {
	RegisterRender(&respXmlRender{})
	RegisterRender(&respYamlRender{})
	RegisterRender(&respJsonRender{})
	RegisterRender(&protoBufRender{})
}

func RegisterRender(r IRender) {
	renders = slice.AppendIfAbsent(renders, r)
}

func MakeRender(mode ESRenderMode) IRender {
	for _, r := range renders {
		if r.Mode() == mode {
			return r
		}
	}
	return nil
}

// ESRender_Xml
type respXmlRender struct{}

func (respXmlRender) Mode() ESRenderMode { return ESRender_Xml }
func (respXmlRender) Render(c *gin.Context, status int, rsp *Response) {
	c.XML(status, rsp)
}

// ESRender_Yaml
type respYamlRender struct{}

func (respYamlRender) Mode() ESRenderMode { return ESRender_Yaml }
func (respYamlRender) Render(c *gin.Context, status int, rsp *Response) {
	c.YAML(status, rsp)
}

// ESRender_Json
type respJsonRender struct{}

func (respJsonRender) Mode() ESRenderMode { return ESRender_Json }
func (respJsonRender) Render(c *gin.Context, status int, rsp *Response) {
	c.PureJSON(status, rsp)
}

// ESRender_Pbuf
type protoBufRender struct{}

func (protoBufRender) Mode() ESRenderMode { return ESRender_Pbuf }
func (protoBufRender) Render(c *gin.Context, status int, rsp *Response) {
	resp := &pb.Response{
		Code:  int32(rsp.Code),
		Msg:   rsp.Msg,
		Error: rsp.Error,
	}
	if rsp.Data != nil {
		anymsg, err := anypb.New(rsp.Data.(proto.Message))
		if err != nil {
			panic(err)
		}
		resp.Data = anymsg
	}
	c.ProtoBuf(status, resp)
}
