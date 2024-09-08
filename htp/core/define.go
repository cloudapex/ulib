package core

import "fmt"

// http heads keys
const (
	C_HTTP_HEAD_REQ_ID       = "c-request-id" // 标示client的每个request的唯一id
	C_HTTP_HEAD_RETRY_AT     = "c-retry-at"   // 如果client发起的request是retry行为,则标识为retry时的时间戳
	C_HTTP_HEAD_CONTENT_TYPE = "Content-Type"
	C_HTTP_HEAD_LANGUAGE     = "c-client-language"
)

// context fields keys
const (
	C_CTX_USER_ID   = "_user_id"   // user_id 字段
	C_CTX_USER_NAME = "_user_name" // user_name 字段
	C_CTX_REQUEST   = "_request"   // Request 字段
	C_CTX_RESPONSE  = "_response"  // Response 字段
	C_CTX_REMOTE_IP = "_remote_ip" // remote_ip 字段
)

// uid类型别名
type TUserID = string // 由于不知道业务层到底是用 string or int64
func IsZeroUID(uid TUserID) bool {
	s := fmt.Sprintf("%v", uid)
	return s == "" || s == "0"
}
