// Package client TODO
package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/def"
	requestor "github.com/coomp/ccs/lib/Requestor"
	"github.com/coomp/ccs/lib/request"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
	"github.com/golang/snappy"
	"github.com/vmihailenco/msgpack"
)

const (
	maxRspDataLen = 65536 // 64k
	// retryTimesWhenUDPCheckFail = 1     // udp验包失败的重试次数，防止串包,野包  暂时不用
)

var uint64seq uint64 = 1
var uint32Seq uint32 = 1
var bufPool = sync.Pool{New: func() interface{} { return make([]byte, maxRspDataLen) }}

// NewUint32Seq 生成全局唯一的uint32 seq
func NewUint32Seq() uint32 {
	return atomic.AddUint32(&uint32Seq, 1)
}

// NewUint64Seq 生成全局唯一的uint64 seq
func NewUint64Seq() uint64 {
	return atomic.AddUint64(&uint64seq, 1)
}

// Codec Codec
type Codec struct {
	CodecType int
	Compress  bool // 是否压缩
}

// Client Client
type Client struct {
	request.Request
	ReqBody *comm.RequestWrapper
	RspBody *comm.ResponseWrapper
	Configs configs.Config
	Codec
}

// New New
func New(addr string, timeout time.Duration, codecType int, compress bool) *Client {
	cli := &Client{
		Request: request.Request{
			Address: addr,
			Network: "tcp",
			ReqType: def.SendAndRecvKeepalive,
			Timeout: timeout,
			Prefix:  "RabbitMQ",
		},
		Codec: Codec{
			CodecType: codecType,
			Compress:  compress,
		},
	}
	return cli
}

// Marshal TODO
func (codec *Client) Marshal() ([]byte, error) {
	pkg := codec.reqBody
	var b []byte
	var e error
	switch codec.CodecType {
	case 4:
		b, e = json.Marshal(pkg)
	case 5:
		b, e = msgpack.Marshal(pkg)
	default:
		e = errors.New(" unsupport codec type")
	}
	if e != nil {
		return nil, e
	}
	var buf []byte
	if codec.Compress {
		b = snappy.Encode(nil, b)
		buf = make([]byte, len(b)+6)
		binary.BigEndian.PutUint32(buf, uint32(len(b)+2))
		buf[4] = 100
		buf[5] = byte(codec.CodecType)
		copy(buf[6:], b)
	} else {
		buf = make([]byte, len(b)+5)
		binary.BigEndian.PutUint32(buf, uint32(len(b)+1))
		buf[4] = byte(codec.CodecType)
		copy(buf[5:], b)
	}
	return buf, nil
}

// Check Check
func (c *Client) Check(data []byte) (int, error) {
	dataLen := len(data)
	if dataLen < 4 {
		return 0, nil
	}
	totalLen := binary.BigEndian.Uint32(data)

	if dataLen < int(totalLen)+4 {
		return 0, nil
	}
	return int(totalLen) + 4, nil

}

// Finish 上报,这里补充下prometheus explore
func (c *Client) Finish(errcode int, address string, cost time.Duration) {
	fmt.Sprintf("errcode:%d_address:%s_cost:%d", errcode, address, cost)
	return
}

// GetInfoFromDataSourceName 拆解下DataSourceName 字段得出
func (c *Client) GetInfoFromDataSourceName(errcode int, address string, cost time.Duration) *requestor.ReqInfo {
	// 明确的是,我们这边需要什么
	// 首先是ip/端口需要从这里拿,其次是网络协议tcp/udp/zmq之类的,再其次超时时间,另外appid,后面用作最终的限流
	fmt.Sprintf("errcode:%d_address:%s_cost:%d", errcode, address, cost)
	return nil
}

// Unmarshal Decode decode a  message to simplessoparser message
func (codec *Client) Unmarshal(b []byte) error {
	if len(b) < 4 {
		return io.ErrUnexpectedEOF
	}
	length := binary.BigEndian.Uint32(b)
	if len(b) < int(length)+4 {
		return io.ErrUnexpectedEOF
	}
	var c byte
	var data []byte
	if b[4] == 100 {
		c = b[5]
		data, _ = snappy.Decode(nil, b[6:])
	} else {
		c = b[4]
		data = b[5:]
	}
	switch c {
	case def.JsonType:
		res := &comm.ResponseWrapper{ResponseData: &comm.Object{Value: &comm.RpcResponse{}}}
		e := json.Unmarshal(data, res)
		codec.rspBody = res
		return e
	case def.Msgpack:
		res := &comm.ResponseWrapper{ResponseData: &comm.Object{Value: &comm.RpcResponse{}}}
		e := msgpack.Unmarshal(data, res)
		codec.rspBody = res
		return e
	default:
		return errors.New("unsupported codec")
	}

}

// DoRequests 多并发请求
func (c *Client) DoRequests(ctx context.Context, reqs ...requestor.Requestor) {
	done := requestor.IsDone(ctx)

	if len(reqs) == 1 {
		req := reqs[0]
		if done > 0 {
			requestor.Finish(req, done, "", 0)
			return
		}
		// 这里要考虑下DataSourceName的设计
		reqInfo := NewReqInfoFromDSN(req.DataSourceName())
		c, f := context.WithTimeout(ctx, reqInfo.Timeout)
		requestor.DoRequest(c, req, reqInfo)
		f()
	} else {
		var wg sync.WaitGroup
		for _, req := range reqs {
			if done > 0 {
				requestor.Finish(req, done, "", 0)
				return
			}

			wg.Add(1)
			reqInfo := NewReqInfoFromDSN(req.DataSourceName())
			subCtx, cancel := context.WithTimeout(ctx, reqInfo.Timeout)
			go func(r requestor, c context.Context, f context.CancelFunc, info *ReqInfo) {
				requestor.DoRequest(c, r, info)
				f()
				wg.Done()
			}(req, subCtx, cancel, reqInfo)
		}
		wg.Wait()
	}
}
