package htp

import (
	"encoding/json"
	"fmt"

	"github.com/cloudapex/ulib/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// > 错误码类型
type ECode int //
const (
	ECodeSucessed  ECode = 0   // http.StatusOK
	ECodeSysError  ECode = 501 // 系统错误
	ECodeParamErr  ECode = 502 // 参数错误
	ECodeMDBError  ECode = 503 // 数据库错误
	ECodeRDBError  ECode = 504 // 缓存库错误
	ECodeCodeCrypt ECode = 505 // 编码加密错误
	ECodeLogicErr  ECode = 506 // 逻辑错误

	ECodeExtendBegin1000 ECode = 1000 // 业务扩展起始编号
) // Inherit from fmt.Stringer interface
func (e ECode) String() string {
	switch e {
	case ECodeSucessed:
		return "ECodeSucessed"
	case ECodeSysError:
		return "ECodeSysError"
	case ECodeParamErr:
		return "ECodeParamErr"
	case ECodeMDBError:
		return "ECodeMDBError"
	case ECodeRDBError:
		return "ECodeRDBError"
	case ECodeCodeCrypt:
		return "ECodeCodeCrypt"
	case ECodeLogicErr:
		return "ECodeLogicErr"
	}
	return fmt.Sprintf("ECode(%d)", e)
}

// Response 基础序列化器
type Response struct {
	Code  int    `json:"code" xml:"code" yaml:"code"`
	Data  any    `json:"data,omitempty" xml:"data,omitempty" yaml:"data,omitempty"`
	Msg   string `json:"msg" xml:"msg" yaml:"msg"`
	Error string `json:"err,omitempty" xml:"err,omitempty" yaml:"err,omitempty"` // code_string : error_info

	file bool // 程序内部使用
}

// 是否成功处理请求
func (r *Response) Sucessed() bool {
	return r.Code == int(ECodeSucessed) || (r.Code != 0 && r.Data != nil)
}

// 成功返回
func RespOK(msg string, data any) Response {
	if msg == "" {
		msg = "ok"
	}
	return Response{Code: int(ECodeSucessed), Data: data, Msg: msg}
}

// 文件返回
func RespFile(c *gin.Context, filePath string) Response {
	c.File(filePath)
	return Response{Code: int(ECodeSucessed), Data: nil, Msg: "ok", file: true}
}

// 通用错误处理
func RespErr[T util.IntStringer](code T, msg string, err error) Response {
	return Response{Code: int(code), Msg: msg, Error: fmt.Sprintf("%v:%v", code, err)}
}

// 通用错误with数据
func RespData[T util.IntStringer](code T, msg string, data any) Response {
	return Response{Code: int(code), Msg: msg, Data: data, Error: fmt.Sprintf("%v:nil", code)}
}

// 数据库操作失败
func RespSysErr(msg string, err error) Response {
	util.Cast(msg == "", func() { msg = "system inter error" }, nil)
	return RespErr(ECodeSysError, msg, err)
}

// 数据库操作失败
func RespMDBErr(msg string, err error) Response {
	util.Cast(msg == "", func() { msg = "mdb operate error" }, nil)
	return RespErr(ECodeMDBError, msg, err)
}

// 缓存库操作失败
func RespRDBErr(msg string, err error) Response {
	util.Cast(msg == "", func() { msg = "rdb operate error" }, nil)
	return RespErr(ECodeRDBError, msg, err)
}

// 绑定器返回错误
// https://github.com/go-playground/validator/blob/master/_examples/simple/main.go
func RespBindErr(err error) Response {
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := fmt.Sprintf("Field.%s", e.Field())
			tag := fmt.Sprintf("Tag.Valid.%s", e.Tag())
			return RespParamErr(fmt.Sprintf("%s%s", field, tag), err)
		}
	}
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return RespParamErr("field type mismatch", err)
	}

	return RespParamErr("", err)
}

// 各种参数错误
func RespParamErr(msg string, err error) Response {
	util.Cast(msg == "", func() { msg = "param error" }, nil)
	return RespErr(ECodeParamErr, msg, err)
}

// 功能逻辑错误
func RespLogicErr(msg string, err error) Response {
	util.Cast(msg == "", func() { msg = "logic error" }, nil)
	return RespErr(ECodeLogicErr, msg, err)
}
