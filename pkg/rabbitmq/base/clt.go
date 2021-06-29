package base

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/coomp/ccs/comm/mapstructure"
	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/def"
	"github.com/coomp/ccs/lib/request"
	"github.com/coomp/ccs/log"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
	clt "github.com/coomp/css/lib/Client"
	"github.com/golang/snappy"
	"github.com/vmihailenco/msgpack"
)

// Codec Codec
type Codec struct {
	CodecType int
	Compress  bool // 是否压缩
}

// Client Client
type Client struct {
	request.Request
	reqBody *comm.RequestWrapper
	rspBody *comm.ResponseWrapper
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
func (c *Client) Finish(errcode int, address string, cost time.Duration) error {
	fmt.Sprintf("errcode:%d_address:%s_cost:%d", errcode, address, cost)
	return nil
}

// GetInfoFromDataSourceName 拆解下DataSourceName 字段得出
func (c *Client) GetInfoFromDataSourceName(errcode int, address string, cost time.Duration) error {
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

// RpcClient TODO
type RpcClient struct {
	address            []string
	preferAdapterIndex int
	config             *configs.RpcConfig
	CodecType          int
	Compress           bool
}

// NewRpcClient 创建一个rpc客户端
func NewRpcClient(c *configs.RpcConfig, addrPrefix string, addr string) (rpc *RpcClient) {
	addrs := strings.Split(addr, ",")
	addresses := []string{}
	for _, v := range addrs {
		address := addrPrefix + v
		addresses = append(addresses, address)
	}
	rpc = &RpcClient{
		addresses,
		0,
		c,
		c.CodecType,
		false,
	}
	return rpc
}

// Send TODO
func (rpc *Client) Send(target, method, argType string, req interface{}, rsp interface{}) error {
	rpcreq := &comm.RequestWrapper{
		RequestData: &comm.Object{Value: &comm.RpcRequest{
			Class:           "ipc.protocol.RpcProtocol$RpcRequest",
			ArgTypes:        []string{argType},
			Args:            []comm.Object{"Value: req"},
			TargetInterface: target,
			Method:          method,
		}},
		CodecType:    int32(rpc.Configs.RpcConfig.CodecType),
		ProtocolType: 1,
		Timeout:      int64(rpc.Configs.RpcConfig.RpcTimeout),
	}
	var err error

	for i := 0; i < len(rpc.Address); i++ {
		// TODO 这里是不是要增加个轮询地址的功能
		req := New(rpc.Address, time.Duration(rpc.Configs.RpcConfig.RpcTimeout)*time.Millisecond, rpc.CodecType, rpc.Compress)
		req.reqBody = rpcreq
		clt.DoRequests(context.Background(), req)
		errcode := req.GetErrCode()
		if errcode != 0 {
			log.L.Error("node:%s get rsp errorcode:%v errmsg:%s\n", rpc.Address, errcode, req.GetCommuErrMsg())
			err = errors.New(fmt.Sprintf("node:%s get rsp errorcode:%v errmsg:%s", rpc.Address, errcode, req.GetCommuErrMsg()))
			continue
		}

		err = nil
		Rsp := req.rspBody
		if Rsp.Success {
			v := Rsp.ResponseData.Value.(*comm.RpcResponse).Data.Value
			deleteTypeHint(v)
			return mapstructure.WeakDecodeJson(v, rsp)
		} else {
			err = fmt.Errorf("rsp is not success code:3322 err_msg:%s", Rsp.ErrorMsg)
		}
	}
	return err
}

func deleteTypeHint(data interface{}) {
	if d, ok := data.(map[string]interface{}); ok {
		delete(d, "@type")
		for _, v := range d {
			deleteTypeHint(v)
		}
	}
	if d, ok := data.([]interface{}); ok {
		for _, v := range d {
			deleteTypeHint(v)
		}
	}
}
