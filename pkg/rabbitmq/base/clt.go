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
	appid              string             // 期望租户分配的appid
	preferAdapterIndex string             // 期望的连接方式
	Config             *configs.RpcConfig // 也可以通过配置的方式走
	CodecType          int
	Compress           bool
	address            []string // 租户分来的
}

// NewRabbitmqRpcClient 创建一个Rabbitmqrpc客户端,底层使用client去访问
func NewRabbitmqRpcClient(c *configs.RpcConfig, appid string) (rpc *RpcClient) {
	//  这里可以通过租户给,也可以通过配置给
	// 租户给 TODO
	// 配置给
	addrs := strings.Split(c.Address, ",")
	addresses := []string{}
	for _, v := range addrs {
		addresses = append(addresses, v)
	}
	rpc = &RpcClient{
		appid,
		"tcp",
		c,
		c.CodecType,
		false,
		addresses,
	}
	return rpc
}

// Send TODO
func (rpc *RpcClient) Send(target, method, argType string, req interface{}, rsp interface{}) error {
	rpcreq := &comm.RequestWrapper{
		RequestData: &comm.Object{Value: &comm.RpcRequest{
			Class:           "ipc.protocol.RpcProtocol$RpcRequest",
			ArgTypes:        []string{argType},
			Args:            []comm.Object{comm.Object{Value: req}},
			TargetInterface: target,
			Method:          method,
		}},
		CodecType:    int32(rpc.Config.CodecType),
		ProtocolType: 1,
		Timeout:      int64(rpc.Config.RpcTimeout),
	}
	var err error
	//

	// for i := 0; i < len(rpc.address); i++ {
	// TODO 这里是不是要增加个轮询地址的功能
	add := ""
	if len(rpc.address) > 0 {
		add = rpc.address[0]
	}

	cReq := client.New(add, time.Duration(rpc.Config.RpcTimeout)*time.Millisecond, rpc.CodecType, rpc.Compress)
	cReq.ReqBody = rpcreq
	cReq.DoRequests(context.Background(), req)
	errcode := cReq.GetErrCode()
	if errcode != 0 {
		log.L.Error("get rsp errorcode:%v errmsg:%s\n", errcode, cReq.GetCommuErrMsg())
		err = errors.New(fmt.Sprintf(" get rsp errorcode:%v errmsg:%s", errcode, cReq.GetCommuErrMsg()))
		continue
	}

	err = nil
	Rsp := cReq.RspBody
	if Rsp.Success {
		v := Rsp.ResponseData.Value.(*comm.RpcResponse).Data.Value
		deleteTypeHint(v)
		return mapstructure.WeakDecodeJson(v, rsp)
	} else {
		err = fmt.Errorf("rsp is not success code:3322 err_msg:%s", Rsp.ErrorMsg)
	}
	//}
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
