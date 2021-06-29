package requestor

import (
	"context"
	"runtime"
	"time"

	"github.com/coomp/ccs/comm"
	"github.com/coomp/ccs/errors"
	"github.com/coomp/ccs/lib/pool"
	"github.com/coomp/ccs/log"
)

// Requestor 后端请求需要实现的接口 an interface that client uses to marshal/unmarshal, and then request
type Requestor interface {
	GetInfoFromDataSourceName(
		errcode int, address string, cost time.Duration) *ReqInfo // DataSourceName //tenant2appid?timeout=300&reqtype=1&network=tcp(tcp/zmq)
	Marshal() ([]byte, error)
	Check([]byte) (int, error)
	Unmarshal([]byte) error
	Finish(errcode int, address string,
		cost time.Duration) // Finish return error code, address, cost time when request finish for report
}

// ReqInfo 后端请求必需信息 由DataSourceName解析出来
type ReqInfo struct {
	Network string        // tcp udp  zmq
	Address string        // ip://ip:port  dns://id.qq.com:80
	ReqType int           // request type: SendAndRecv SendAndRecvKeepalive SendOnlyKeepalive SendOnly SendAndRecvIgnoreError
	Timeout time.Duration // current action timeout time.Second
	ZmqNet  string        // zmq only: tcp inproc
}

// GetInfoFromDataSourceName appid?timeout=300&reqtype=1&network=udp 中领取所有的信息
func GetInfoFromDataSourceName(req Requestor) *ReqInfo {
	return req.GetInfoFromDataSourceName(req)
}

// IsDone 这里暂时先透传出去,这里其实可以让用户提供数据,这里统一上报
func IsDone(ctx context.Context) int {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled {
			return errors.ContextCanceled.Int()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return errors.ContextTimeout.Int()
		}
		return 0
	default:
	}
	return 0
}

// doNetworkRequest
func doNetworkRequest(ctx context.Context, r Requestor, arr string, reqInfo *ReqInfo) int {
	// 网络库
	reqInfo.Address
	pool.GetTCPConnectionPool()
	return 0
}

// Finish TODO
func Finish(req Requestor, errcode int, address string, cost time.Duration) {
	req.Finish(errcode, address, cost)
}

// DoRequest TODO
func DoRequest(ctx context.Context, r Requestor, reqInfo *ReqInfo) {
	s := time.Now()
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 16*1024*1024)
			buf = buf[:runtime.Stack(buf, false)]
			Finish(r, errors.ErrRequestPanic.Int(), "", 0)
		}
	}()

	addr := reqInfo.Address
	// check if done after addressing
	if done := IsDone(ctx); done > 0 {
		Finish(r, done, addr, time.Duration(comm.Timediffer(s)))
		return
	}

	// 下面的分支都需要补充上报
	var ec int
	if reqInfo.Network == "tcp" {
		ec = doNetworkRequest(ctx, r, addr, reqInfo)
	} else if reqInfo.Network == "zmq" {
		// 暂时没有实现，这个zmq还是很强的，使用一个开源库就可以解决问题
	} else {
		Finish(r, errors.ErrNetworkInvalid.Int(), addr, time.Duration(comm.Timediffer(s)))
		return
	}
	Finish(r, ec, addr, time.Duration(comm.Timediffer(s)))

	if ec == errors.ErrDialConnFail.Int() || ec == errors.ErrRecvTimeout.Int() || ec == errors.ErrRecvFail.Int() {
		log.L.Error("DoRequest [%d,%s]", ec, errors.ErrCode(ec).String())
	}
}
