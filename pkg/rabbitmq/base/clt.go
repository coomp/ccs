package base

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coomp/ccs/comm/mapstructure"
	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/lib/client"
	"github.com/coomp/ccs/log"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
)

// RpcClient TODO
type RpcClient struct {
	address            []string
	preferAdapterIndex int
	Config             *configs.RpcConfig
	CodecType          int
	Compress           bool
}

// NewRabbitmqRpcClient 创建一个Rabbitmqrpc客户端,底层使用client去访问
func NewRabbitmqRpcClient(c *configs.RpcConfig, addrPrefix string, addr string) (rpc *RpcClient) {
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
func (rpc *RpcClient) Send(target, method, argType string, req interface{}, rsp interface{}) error {
	rpcreq := &comm.RequestWrapper{
		RequestData: &comm.Object{Value: &comm.RpcRequest{
			Class:           "ipc.protocol.RpcProtocol$RpcRequest",
			ArgTypes:        []string{argType},
			Args:            []comm.Object{"Value: req"},
			TargetInterface: target,
			Method:          method,
		}},
		CodecType:    int32(rpc.Config.CodecType),
		ProtocolType: 1,
		Timeout:      int64(rpc.Config.RpcTimeout),
	}
	var err error

	for i := 0; i < len(rpc.Address); i++ {
		// TODO 这里是不是要增加个轮询地址的功能
		req := client.New(rpc.Address, time.Duration(rpc.Config.RpcTimeout)*time.Millisecond, rpc.CodecType, rpc.Compress)
		req.ReqBody = rpcreq
		req.DoRequests(context.Background(), req)
		errcode := req.GetErrCode()
		if errcode != 0 {
			log.L.Error("get rsp errorcode:%v errmsg:%s\n", errcode, req.GetCommuErrMsg())
			err = errors.New(fmt.Sprintf(" get rsp errorcode:%v errmsg:%s", errcode, req.GetCommuErrMsg()))
			continue
		}

		err = nil
		Rsp := req.RspBody
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
