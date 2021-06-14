package net

import (
	"sync"
	"sync/atomic"
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