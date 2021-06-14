package base

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coomp/ccs/def"
	"github.com/coomp/ccs/net/request"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
	"github.com/golang/snappy"
	"github.com/vmihailenco/msgpack"
	"io"
	"strings"
	"time"
)

type Codec struct {
	CodecType int
	Compress  bool
}

type Client struct {
	request.Request
	reqBody *comm.RequestWrapper
	rspBody *comm.ResponseWrapper
	Codec
}

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

//Encode encode
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
		e = errors.New("hippo unsupport codec type")
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

//Decode decode a hippo message to simplessoparser message
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
	case 4:
		res := &comm.ResponseWrapper{ResponseData: &comm.Object{Value: &comm.RpcResponse{}}}
		e := json.Unmarshal(data, res)
		codec.rspBody = res
		return e
	case 5:
		res := &comm.ResponseWrapper{ResponseData: &comm.Object{Value: &comm.RpcResponse{}}}
		e := msgpack.Unmarshal(data, res)
		codec.rspBody = res
		return e
	default:
		return errors.New("unsupported codec")
	}

}

type RpcClient struct {
	address            []string
	preferAdapterIndex int
	config             *config.RpcConfig
	CodecType          int
	Compress           bool
}

func NewHippoRpcClient(c *config.RpcConfig, addrPrefix string, addr string, log *config.LogWrapper) (rpc *RpcClient) {

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
		c.Compress,
	}
	return rpc
}

func (rpc *HippoRpcClient) Send(ctx context.Context, target, method, argType string, req interface{}, rsp interface{}) error {

	rpcreq := &def.RequestWrapper{
		Class: "com.tencent.hippo.ipc.RequestWrapper",
		RequestData: &def.Object{Value: &def.RpcRequest{
			Class:           "com.tencent.hippo.ipc.protocol.RpcProtocol$RpcRequest",
			ArgTypes:        []string{argType},
			Args:            []def.Object{def.Object{Value: req}},
			TargetInterface: target,
			Method:          method,
		}},
		CodecType:    int32(rpc.config.CodecType),
		ProtocolType: 1,
		Timeout:      int64(rpc.config.RpcTimeout),
	}
	var err error

	for i := 0; i < len(rpc.address); i++ {
		//轮询地址
		index := (rpc.preferAdapterIndex + i) % len(rpc.address)
		req := New(rpc.address[index], time.Duration(rpc.config.RpcTimeout)*time.Millisecond, rpc.CodecType, rpc.Compress)
		req.reqBody = rpcreq
		client.DoRequests(context.Background(), req)

		errcode := req.GetErrCode()

		if errcode != 0 {
			attr.AttrAPI(monitor.HIPPO_CLIENT_DO_REQ_FAIL, 1) //[HippoClient]doRequest失败量
			//一个节点不可用，继续循环下个节点。否则配多个节点没有意义
			rpc.logger.Errorf("node:%s get rsp errorcode:%v errmsg:%s\n", rpc.address[index], errcode, req.GetCommuErrMsg())
			err = errors.New(fmt.Sprintf("node:%s get rsp errorcode:%v errmsg:%s", rpc.address[index], errcode, req.GetCommuErrMsg()))
			continue
		}
		err = nil
		hippoRsp := req.rspBody
		if hippoRsp.Success {
			rpc.preferAdapterIndex = index //成功则下次还是用这个node
			v := hippoRsp.ResponseData.Value.(*def.RpcResponse).Data.Value
			deleteTypeHint(v)
			return mapstructure.WeakDecodeJson(v, rsp)
		} else {
			attr.AttrAPI(monitor.HIPPO_CLIENT_RSP_FAIL, 1) //[HippoClient]返回的rsp不为success
			err = fmt.Errorf("hippo_rsp is not success code:3322 err_msg:%s", hippoRsp.ErrorMsg)
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

