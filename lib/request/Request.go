package request

import (
	"github.com/coomp/ccs/errors"
	"time"
)

// Request Requestor接口的具体实现
type Request struct {
	ReqType int           // request type: SendAndRecv SendAndRecvKeepalive SendOnlyKeepalive SendOnly SendAndRecvIgnoreError
	Network string        // tcp udp unix
	Address string        // ip://ip:port  dns://id.qq.com:80
	Timeout time.Duration // current action timeout time.Second

	ErrCode int           // return error code after finish
	IPPort  string        // return ip:port address after addressing
	Cost    time.Duration // return cost time after finish

	Command        string // service request command name, for jm report
	Prefix         string // for jm report
	Sequence       uint32 // service packet sequence
	ServiceErrCode int    // for monitor
	ServiceErrMsg  string // for monitor
}

// GetErrCode get final error code
func (r *Request) GetErrCode() int {
	if r.ErrCode != 0 {
		return r.ErrCode
	} else if r.ServiceErrCode != 0 {
		return r.ServiceErrCode
	}
	return 0
}


// GetCommuErrMsg get commu error message
func (r *Request) GetCommuErrMsg() string {
	return errors.ErrCode(r.ErrCode).String()
}