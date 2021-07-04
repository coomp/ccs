package requestor

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/coomp/ccs/comm"
	"github.com/coomp/ccs/def"
	"github.com/coomp/ccs/errors"
	"github.com/coomp/ccs/lib/pool"
	"github.com/coomp/ccs/log"
)

// Requestor 后端请求需要实现的接口 an interface that client uses to marshal/unmarshal, and then request
type Requestor interface {
	GetInfoFromDataSourceName() (*ReqInfo,
		error) // DataSourceName //tenant2appid?timeout=300&reqtype=1&network=tcp(tcp/zmq)
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
	Appid   int           // appid
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
func doNetworkRequest(ctx context.Context, r Requestor, reqInfo *ReqInfo) int {
	// 网络库
	d, _ := ctx.Deadline()
	var conn net.Conn
	shouldReturnPool := false
	p := pool.GetTCPConnectionPool(fmt.Sprintf("%s", reqInfo.Appid), reqInfo.Address, reqInfo.Network, reqInfo.Timeout)
	if reqInfo.ReqType == def.SendAndRecvKeepalive || reqInfo.ReqType == def.SendOnlyKeepalive { // 长连接
		shouldReturnPool = true
		c, err := p.Get()
		if err != nil {
			return errors.DialFail.Int()
		}
		conn = c.C

	} else {
		if c, err := net.DialTimeout(reqInfo.Network, reqInfo.Address, reqInfo.Timeout); err != nil {
			return errors.DialFail.Int()
		} else {
			conn = c
		}
	}
	defer func() {
		if shouldReturnPool && p != nil {
			p.Put(conn)
		} else {
			conn.Close()
			if p != nil && !shouldReturnPool {
				//TODO 上报连接失效移出连接池
			}
		}
	}()
	d, _ = ctx.Deadline()
	conn.SetDeadline(d)
	reqData, err := r.Marshal()
	if err != nil || len(reqData) == 0 {
		log.L.Errorf("marshal fail:%s, req data len:%d", err, len(reqData))
		return errors.MarshalFail.Int()
	}
	sentNum := 0
	for sentNum < len(reqData) {
		var num int
		num, err = conn.Write(reqData[sentNum:])
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return errors.SendTimeout.Int()
			}
			log.L.Errorf("send fail:%v", err)
			return errors.SendFail.Int()
		}
		sentNum += num

		// check if done after write
		if done := IsDone(ctx); done > 0 {
			return done
		}
		buf := def.BufPool.Get()
		rspData, _ := buf.([]byte)
		defer func() {
			def.BufPool.Put(rspData) //tcp包过大扩充时会重新赋值，所以defer必须放在闭包里面
		}()

		recvNum := 0
		checkNum := 0
		for {
			var num int
			num, err = conn.Read(rspData[recvNum:])
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return errors.ErrRecvTimeout.Int()
				}
				log.L.Errorf("recv fail:%v", err)
				return errors.ErrRecvFail.Int()
			}
			recvNum += num
			if recvNum >= cap(rspData) {
				if recvNum >= 1024*def.MaxRspDataLen {
					fmt.Println("recv rsp data too big, larger than 64M, return fail, recv num:", recvNum)
					return errors.ErrRspDataTooLarge.Int()
				}
				fmt.Println("recv rsp data too big, expand twice cap, recv num:", recvNum)
				tmpRspData := make([]byte, recvNum*2)
				copy(tmpRspData, rspData[:recvNum])
				rspData = tmpRspData
			}

			// check if done after read
			if done := IsDone(ctx); done > 0 {
				return done
			}

			checkNum, err = r.Check(rspData[:recvNum])
			if err != nil || checkNum < 0 {
				return errors.ErrCheckFail.Int()
			}
			if checkNum > 0 {
				if checkNum < recvNum {
				}
				if checkNum > recvNum {
					return errors.ErrCheckFail.Int()
				}
				break
			}
		}

		err = r.Unmarshal(rspData[:checkNum])
		if err != nil {
			log.L.Errorf("unmarshal fail:%s", err)
			return errors.MarshalFail.Int()
		}

		shouldReturnPool = true
		return errors.Succ.Int()
	}
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
		ec = doNetworkRequest(ctx, r, reqInfo)
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
