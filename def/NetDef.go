package def

import "sync"

// 后端请求类型 request type
const (
	SendAndRecv            = 1 // 普通的一来一回req/rsp
	SendAndRecvKeepalive   = 2 // 使用tcp连接池
	SendOnlyKeepalive      = 3 // tcp连接池只发不收
	SendOnly               = 4 // 只发不收
	SendAndRecvIgnoreError = 5 // web长轮询技术
	SendAndRecvStream      = 6 // tcp stream transport
	SendOnlyStream         = 7 // tcp流一直发
)

const (
	// JsonType json 传输协议
	JsonType = 1
	// Msgpack TO msgpack 传输协议DO
	Msgpack = 2
)

const (
	// READY TODO
	READY NodeState = 0
	// RUNNING TODO
	RUNNING NodeState = 1
	// STOP TODO
	STOP NodeState = 2
)

const (
	// MaxRspDataLen TODO
	MaxRspDataLen              = 65536 // 64k
	retryTimesWhenUDPCheckFail = 1     // udp验包失败的重试次数，防止串包,野包
)

// BufPool TODO
var BufPool = sync.Pool{New: func() interface{} { return make([]byte, MaxRspDataLen) }}
