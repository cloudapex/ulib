package log

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

// What compression type the writer should use when sending messages
// to the graylog server
type CompressType int

const (
	CompressGzip CompressType = iota
	CompressZlib
	CompressNone
)

// Used to control GELF chunking.  Should be less than (MTU - len(UDP
// header)).
//
// TODO: generate dynamically using Path MTU Discovery?
const (
	ChunkSize        = 1420
	chunkedHeaderLen = 12
	chunkedDataLen   = ChunkSize - chunkedHeaderLen
)

var (
	magicChunked = []byte{0x1e, 0x0f}
	magicZlib    = []byte{0x78}
	magicGzip    = []byte{0x1f, 0x8b}
)

// 1k bytes buffer by default
var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

func newBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	if b != nil {
		b.Reset()
		return b
	}
	return bytes.NewBuffer(nil)
}

// 安装gelf拦截器,需要首先调用 Init 方法
func UseGelf(conf *GraylogConf) {
	if conf == nil || conf.Address == "" {
		return
	}
	var err error
	hostname, _ := os.Hostname()
	facility, withFull, intercept := path.Base(os.Args[0]), conf.WithFull, conf.GelfIntercept
	if conf.Service != "" {
		facility = conf.Service
	}

	w := new(Writer)
	w.CompressionLevel = flate.BestSpeed

	if w.conn, err = net.Dial("udp", conf.Address); err != nil {
		Error("connect to graylog[%s], err:%v", conf.Address, err)
		return
	}

	Filter(func(msg *LogUnit) bool {
		body := []byte(msg.Str)
		short, full := body, []byte("")
		if withFull {
			if i := bytes.IndexRune(body, '\n'); i > 0 && len(body) > i+1 {
				short, full = body[:i], body
			}
		}

		m := Message{
			Version:     "1.1",
			Host:        hostname,
			Facility:    facility,
			TimeUnix:    float64(msg.At.UnixNano()/1e6) / 1000,
			CreateOrder: w.calCreateOrder(msg.At),
			Level:       int32(msg.Lv),
			LevelName:   msg.Lv.String(),
			Short:       string(short),
			Full:        string(full),
			MsgSize:     fmt.Sprintf("%0.2fk", float32(len(short))/1024),
			Extra:       msg.Fields,
		}

		if err := w.Write(&m); err != nil {
			go func() {
				at := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("%v ulog output err:%v, with content:%s", at, err, string(short[0:int(math.Min(100, float64(len(short))))]))
			}()
		}
		return intercept
	})
}

// Writer implements io.Writer and is used to send both discrete
// messages to a graylog2 server, or data from a stream-oriented
// interface (like the functions in log).
type Writer struct {
	mu               sync.Mutex
	conn             net.Conn
	hostname         string
	Facility         string // defaults to current process name
	CompressionLevel int    // one of the consts from compress/flate
	CompressionType  CompressType
	lastSecond       int64
	IncPerSecond     int
}

// Message represents the contents of the GELF message.  It is gzipped
// before sending.
type Message struct {
	Version     string                 `json:"version"`                // internal
	Host        string                 `json:"host"`                   // internal
	Short       string                 `json:"short_message"`          // internal(message field)
	Full        string                 `json:"full_message,omitempty"` // invalid
	TimeUnix    float64                `json:"timestamp"`              // 时间戳
	CreateOrder int64                  `json:"message_order"`          // 消息生产顺序
	Level       int32                  `json:"level"`                  // 日志等级
	LevelName   string                 `json:"level_name,omitempty"`   // 日志等级名称
	Facility    string                 `json:"tag,omitempty"`          // 服务名称
	MsgSize     string                 `json:"message_size,omitempty"` // 消息大小
	Extra       map[string]interface{} `json:"-"`
	RawExtra    json.RawMessage        `json:"-"` // 扩展用
}

// WriteMessage sends the specified message to the GELF server
// specified in the call to New().  It assumes all the fields are
// filled out appropriately.  In general, clients will want to use
// Write, rather than WriteMessage.
func (w *Writer) Write(m *Message) (err error) {
	mBuf := newBuffer()
	defer bufPool.Put(mBuf)
	if err = m.marshalJsonBuf(mBuf); err != nil {
		return err
	}
	mBytes := mBuf.Bytes()

	var (
		zBuf   *bytes.Buffer
		zBytes []byte
	)

	var zw io.WriteCloser
	switch w.CompressionType {
	case CompressGzip:
		zBuf = newBuffer()
		defer bufPool.Put(zBuf)
		zw, err = gzip.NewWriterLevel(zBuf, w.CompressionLevel)
	case CompressZlib:
		zBuf = newBuffer()
		defer bufPool.Put(zBuf)
		zw, err = zlib.NewWriterLevel(zBuf, w.CompressionLevel)
	case CompressNone:
		zBytes = mBytes
	default:
		panic(fmt.Sprintf("unknown compression type %d",
			w.CompressionType))
	}
	if zw != nil {
		if err != nil {
			return
		}
		if _, err = zw.Write(mBytes); err != nil {
			zw.Close()
			return
		}
		zw.Close()
		zBytes = zBuf.Bytes()
	}

	if numChunks(zBytes) > 1 {
		return w.writeChunked(zBytes)
	}
	n, err := w.conn.Write(zBytes)
	if err != nil {
		return
	}
	if n != len(zBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(zBytes))
	}

	return nil
}

// Close connection and interrupt blocked Read or Write operations
func (w *Writer) Close() error {
	return w.conn.Close()
}

// writes the gzip compressed byte array to the connection as a series
// of GELF chunked messages.  The format is documented at
// http://docs.graylog.org/en/2.1/pages/gelf.html as:
//
//	2-byte magic (0x1e 0x0f), 8 byte id, 1 byte sequence id, 1 byte
//	total, chunk-data
func (w *Writer) writeChunked(zBytes []byte) (err error) {
	b := make([]byte, 0, ChunkSize)
	buf := bytes.NewBuffer(b)
	nChunksI := numChunks(zBytes)
	if nChunksI > 128 {
		return fmt.Errorf("msg too large, would need %d chunks", nChunksI)
	}
	nChunks := uint8(nChunksI)
	// use urandom to get a unique message id
	msgId := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, msgId)
	if err != nil || n != 8 {
		return fmt.Errorf("rand.Reader: %d/%s", n, err)
	}

	bytesLeft := len(zBytes)
	for i := uint8(0); i < nChunks; i++ {
		buf.Reset()
		// manually write header.  Don't care about
		// host/network byte order, because the spec only
		// deals in individual bytes.
		buf.Write(magicChunked) //magic
		buf.Write(msgId)
		buf.WriteByte(i)
		buf.WriteByte(nChunks)
		// slice out our chunk from zBytes
		chunkLen := chunkedDataLen
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}
		off := int(i) * chunkedDataLen
		chunk := zBytes[off : off+chunkLen]
		buf.Write(chunk)

		// write this chunk, and make sure the write was good
		n, err := w.conn.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("Write (chunk %d/%d): %s", i,
				nChunks, err)
		}
		if n != len(buf.Bytes()) {
			return fmt.Errorf("Write len: (chunk %d/%d) (%d/%d)",
				i, nChunks, n, len(buf.Bytes()))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}
	return nil
}
func (w *Writer) calCreateOrder(at time.Time) int64 {
	if w.lastSecond == at.Unix() {
		if w.IncPerSecond++; w.IncPerSecond > 9999 {
			w.IncPerSecond = 9999
		}
	} else {
		w.lastSecond = at.Unix()
		w.IncPerSecond = 0
	}
	return w.lastSecond*10000 + int64(w.IncPerSecond)
}
func (m *Message) marshalJsonBuf(buf *bytes.Buffer) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// write up until the final }
	if _, err = buf.Write(b[:len(b)-1]); err != nil {
		return err
	}
	if len(m.Extra) > 0 {
		eb, err := json.Marshal(m.Extra)
		if err != nil {
			return err
		}
		// merge serialized message + serialized extra map
		if err = buf.WriteByte(','); err != nil {
			return err
		}
		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(eb[1 : len(eb)-1]); err != nil {
			return err
		}
	}

	if len(m.RawExtra) > 0 {
		if err := buf.WriteByte(','); err != nil {
			return err
		}

		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(m.RawExtra[1 : len(m.RawExtra)-1]); err != nil {
			return err
		}
	}

	// write final closing quotes
	return buf.WriteByte('}')
}

// numChunks returns the number of GELF chunks necessary to transmit
// the given compressed buffer.
func numChunks(b []byte) int {
	lenB := len(b)
	if lenB <= ChunkSize {
		return 1
	}
	return len(b)/chunkedDataLen + 1
}
