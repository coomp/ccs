package lib

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coomp/ccs/errors"
)

const (
	maxRspDataLen              = 65536 // 64k
	retryTimesWhenUDPCheckFail = 1     // udp验包失败的重试次数，防止串包,野包
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

// Requestor 后端请求需要实现的接口 an interface that client uses to marshal/unmarshal, and then request
type Requestor interface {
	DataSourceName() string // DataSourceName cmlb://appid?timeout=300&reqtype=1&network=udp
	Cmd() string
	Marshal() ([]byte, error)
	Check([]byte) (int, error)
	Unmarshal([]byte) error
	Finish(errcode int, address string,
		cost time.Duration) // Finish return error code, address, cost time when request finish
}

func isDone(ctx context.Context) int {
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

// DoRequests 多并发请求
func DoRequests(ctx context.Context, reqs ...Requestor) {
	done := isDone(ctx)

	if len(reqs) == 1 {
		req := reqs[0]
		if done > 0 {
			finish(ctx, req, done, "", 0)
			return
		}

		reqInfo := NewReqInfoFromDSN(req.DataSourceName())
		c, f := context.WithTimeout(ctx, reqInfo.Timeout)
		doRequest(c, req, reqInfo)
		f()
	} else {
		var wg sync.WaitGroup
		for _, req := range reqs {
			if done > 0 {
				finish(ctx, req, done, "", 0)
				return
			}

			wg.Add(1)
			reqInfo := NewReqInfoFromDSN(req.DataSourceName())
			subCtx, cancel := context.WithTimeout(ctx, reqInfo.Timeout)
			go func(r Requestor, c context.Context, f context.CancelFunc, info *ReqInfo) {
				doRequest(c, r, info)
				f()
				wg.Done()
			}(req, subCtx, cancel, reqInfo)
		}
		wg.Wait()
	}
}
