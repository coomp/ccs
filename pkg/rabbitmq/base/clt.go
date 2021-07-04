package base

import (
	"context"
	"errors"
	"fmt"

	"github.com/coomp/ccs/comm/mapstructure"
	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/lib/client"
	"github.com/coomp/ccs/log"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
)

// RpcClient rabbitmq的客户端
type RpcClient struct {
	appid              string             // 期望租户分配的appid
	preferAdapterIndex string             // 期望的连接方式
	Config             *configs.RpcConfig // 也可以通过配置的方式走
	Compress           bool
}

// NewRabbitmqRpcClient 创建一个Rabbitmqrpc客户端,底层使用client去访问
func NewRabbitmqRpcClient(c *configs.RpcConfig, appid string) (rpc *RpcClient) {
	//  这里可以通过租户给,也可以通过配置给
	// 租户给 TODO
	// 暂时配置给
	rpc = &RpcClient{
		appid,
		"tcp",
		c,
		false,
	}
	return rpc
}

// Send 发送消息
func (rpc *RpcClient) Send(argType string, req interface{}, rsp interface{}) error {
	rpcreq := &comm.RequestWrapper{
		RequestData: &comm.Object{Value: &comm.RpcRequest{
			Class:    "ipc.protocol.RpcProtocol$RpcRequest",
			ArgTypes: []string{argType},
			Args:     []comm.Object{comm.Object{Value: req}},
		}},
		ProtocolType: 1,
	}
	var err error

	// interface 转换
	cReq := client.New(rpc.Compress)
	cReq.ReqBody = rpcreq
	cReq.DoRequests(context.Background(), cReq)
	errcode := cReq.GetErrCode()
	if errcode != 0 {
		log.L.Error("get rsp errorcode:%v errmsg:%s\n", errcode, cReq.GetCommuErrMsg())
		err = errors.New(fmt.Sprintf(" get rsp errorcode:%v errmsg:%s", errcode, cReq.GetCommuErrMsg()))
		return fmt.Errorf("get_client_err:%d", errcode)
	}

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
