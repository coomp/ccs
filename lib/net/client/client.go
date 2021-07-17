// Package client TODO
package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/def"
	"github.com/coomp/ccs/lib/net/request"
	"github.com/coomp/ccs/lib/net/requestor"
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
	ReqBody    *comm.RequestWrapper
	RspBody    *comm.ResponseWrapper
	Configs    configs.Config // 可以通过config拿dataSource信息
	DataSource string         // 也可以通过租户下发
	Codec
}

// New New
func New(compress bool) *Client {
	cli := &Client{
		Request: request.Request{
			Network: "tcp",
			ReqType: def.SendAndRecvKeepalive,
			Prefix:  "RabbitMQ",
		},
		Codec: Codec{
			Compress: compress,
		},
	}
	return cli
}

// Marshal TODO
func (codec *Client) Marshal() ([]byte, error) {
	pkg := codec.ReqBody
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

// Disassemble  参考xorm的解析，是根据不同的sql引擎进行字符串解析，并不使用正则,正则的性能不强,不过这里每次都仅仅是初始化一次，应该还好
func Disassemble(dataSourceName string) (ReqInfo *requestor.ReqInfo, err error) {
	//appid?timeout=300&reqtype=1&network=udp&address="xxx"
	// 拆解appid和具体的uri
	s := strings.Split(dataSourceName, "?")
	if len(s) < 2 {
		err = fmt.Errorf("dataSourceName have no appid:%s", dataSourceName)
		return
	}
	ReqInfo.Appid, err = strconv.Atoi(s[0])
	if err != nil {
		err = fmt.Errorf("Disassemble atoi err:%s", err.Error())
	}
	if u, perr := comm.ParsePortalMessage(s[1]); perr != nil {
		err = fmt.Errorf("Disassemble ParsePortalMessage err:%s", err.Error())
		return
	} else {
		if itimeout, ierr := strconv.Atoi(u.Get("timeout")); ierr != nil {
			err = fmt.Errorf("Disassemble atoi err:%s", err.Error())
			return
		} else {
			ReqInfo.Timeout = time.Duration(itimeout) * time.Second
		}
		if iReqType, ierr := strconv.Atoi(u.Get("reqtype")); ierr != nil {
			err = fmt.Errorf("Disassemble atoi err:%s", err.Error())
			return
		} else {
			ReqInfo.ReqType = iReqType
		}
		ReqInfo.Address = u.Get("address")
		ReqInfo.Network = u.Get("network")
	}
	return
}

// GetInfoFromDataSourceName 拆解下DataSourceName 字段得出
func (c *Client) GetInfoFromDataSourceName() (ReqInfo *requestor.ReqInfo, err error) {
	// 明确的是,我们这边需要什么
	// 首先是ip/端口需要从这里拿,其次是网络协议tcp/udp/zmq之类的,再其次超时时间,另外appid,后面用作最终的限流
	// appid?timeout=300&reqtype=1&network=udp
	// TODO 查询一下租户,是否从他那边下发，暂时都通过配置走
	ReqInfo, err = Disassemble(c.DataSource)
	return ReqInfo, err
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
		codec.RspBody = res
		return e
	case def.Msgpack:
		res := &comm.ResponseWrapper{ResponseData: &comm.Object{Value: &comm.RpcResponse{}}}
		e := msgpack.Unmarshal(data, res)
		codec.RspBody = res
		return e
	default:
		return errors.New("unsupported codec")
	}

}

// DoRequests 多并发请求
func (c *Client) DoRequests(ctx context.Context, reqs ...requestor.Requestor) error {
	done := requestor.IsDone(ctx)

	if len(reqs) == 1 {
		req := reqs[0]
		if done > 0 {
			requestor.Finish(req, done, "", 0)
			return nil
		}
		// 这里要考虑下DataSourceName的设计
		reqInfo, err := c.GetInfoFromDataSourceName()
		if err != nil {
			return err
		}
		c, f := context.WithTimeout(ctx, reqInfo.Timeout)
		requestor.DoRequest(c, req, reqInfo)
		f()
	} else {
		var wg sync.WaitGroup
		for _, req := range reqs {
			if done > 0 {
				requestor.Finish(req, done, "", 0)
				return nil
			}

			wg.Add(1)
			reqInfo, err := c.GetInfoFromDataSourceName()
			if err != nil {
				return err
			}
			subCtx, cancel := context.WithTimeout(ctx, reqInfo.Timeout)
			go func(r requestor.Requestor, c context.Context, f context.CancelFunc, info *requestor.ReqInfo) {
				requestor.DoRequest(c, r, info)
				f()
				wg.Done()
			}(req, subCtx, cancel, reqInfo)
		}
		wg.Wait()
	}
	return nil
}
