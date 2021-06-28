// Package client TODO
package client

import (
	"context"
	"sync"
	"sync/atomic"

	requestor "github.com/coomp/lib/Requestor"
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

// DoRequests 多并发请求
func DoRequests(ctx context.Context, reqs ...requestor.Requestor) {
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
