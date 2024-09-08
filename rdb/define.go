package rdb

import (
	"errors"
	"time"

	"github.com/cloudapex/ulib/util"

	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

// ==================== 常量定义
const (
	C_CONN_IDLE_TIMEOVER = 5 * time.Minute // 空闲多久回收连接

	C_INF_MAX = "+inf" // 正无穷
	C_INF_MIN = "-inf" // 负无穷

	C_DB_CLUSTER = "_cluster_" // 如果使用集群部署,Config只能有一个且name='_cluster_'

	C_TICK_OUTPUT_STAT_INTERVAL = 1 * time.Minute // 一分钟统计输出一次命令调用情况
)

// ==================== 错误定义
var (
	ErrNil                      = redis.ErrNil
	ErrNameRepeated             = errors.New("ErrConfigNameRepeated")
	ErrInvalidTTL               = errors.New("ErrInvalidTTL")
	ErrZsetAddArrInvalid        = errors.New("ErrZsetAddArrInvalid")
	ErrHashGetMapResultMismatch = errors.New("ErrHashGetMapResultMismatch")
	ErrInvalidCoding            = errors.New("ErrInvalidCoding")
)

// ==================== 类型定义

type ECoding = string // 编码类型
const (
	ECod_None      ECoding = ""
	ECod_Json      ECoding = "json"
	ECod_Msgpack   ECoding = "msgpack"
	ECod_ProtoBuff ECoding = "pb"
)

// > 值的设置模式
type ESetMode int //
const (
	_                ESetMode = iota
	ESet_Update               // 正常设置
	ESet_WhenExist            // 当值存在时才进行设置
	ESet_WhenNoExist          // 当值不存在时才进行设置
	ESet_WhenValueLT          // 当新的分值比当前分值小才设置,不存在则新增
	ESet_WhenValueGT          // 当新的分值比当前分值大才设置,不存在则新增
)

// > 排序类型
type ESort int //
const (
	ESort_Asc  ESort = iota + 1 // 正序 从小到大
	ESort_Desc                  // 倒序 从大到小
)

// > 连接模式
type EMode int //
const (
	EMode_Normal  EMode = iota + 1 // 单库
	EMode_Cluster                  // 集群
)

// 两种连接模式
type NormalPool struct{ redis.Pool } //
func (*NormalPool) Mode() EMode      { return EMode_Normal }

type ClusterPool struct{ redisc.Cluster } //
func (*ClusterPool) Mode() EMode          { return EMode_Cluster }

// > Redis配置
type Config struct {
	Name      string `json:"name"` // 常量 C_DB_CLUSTER 是唯一的
	Addr      string `json:"addr"`
	DbIdx     int64  `json:"index"` // 0-15
	Passwd    string `json:"passwd"`
	MaxIdle   int64  `json:"max_idle"`
	MaxActive int64  `json:"max_active"`
	Wait      bool   `json:"wait"` // 如果当前的连接已经达到MaxActive值的话, 如果true=>线程将会等待, 否则直接报错.
}

// > 用于Sender的连接计数
type sendcc struct {
	conn  redis.Conn // 用于Sender的连接
	count int        // 用于Sender的计数
}

// > Sender
func Sender(k IKeyer, send func(), rep ...func(rps []*Resp)) sender {
	var rp func(rps []*Resp)
	util.Cast(len(rep) != 0, func() { rp = rep[0] }, nil)
	return sender{k.key(), send, rp}
}

type sender struct {
	K     *Key              // 源
	Sends func()            // send 命令...
	Reply func(rps []*Resp) // 回调 命令结果...
}
