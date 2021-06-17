package def

// 后端请求类型 request type
const (
	SendAndRecv            = 1 // 普通的一来一回req/rsp
	SendAndRecvKeepalive   = 2 // 使用tcp连接池
	SendOnlyKeepalive      = 3 // tcp连接池只发不收
	SendOnly               = 4 // 只发不收
	SendAndRecvIgnoreError = 5 // web长轮询技术，没有消息返回超时是正常的，应该忽略不上报l5
	SendAndRecvStream      = 6 // tcp stream transport
	SendOnlyStream         = 7 // tcp流一直发
)

const (
	// JsonType json 传输协议
	JsonType = 1
	// Msgpack TO msgpack 传输协议DO
	Msgpack = 2
)
